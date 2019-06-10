package ssz

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

const lengthBytes = 4

// decodeError is what gets reported to the decoder user in error case.
type decodeError struct {
	msg string
	typ reflect.Type
}

func newDecodeError(msg string, typ reflect.Type) *decodeError {
	return &decodeError{msg, typ}
}

func (err *decodeError) Error() string {
	return fmt.Sprintf("decode error: %s for output type %v", err.msg, err.typ)
}

// Decode SSZ encoded data and output it into the object pointed by pointer val.
func Decode(input []byte, val interface{}) error {
	if val == nil {
		return newDecodeError("cannot decode into nil", nil)
	}
	rval := reflect.ValueOf(val)
	rtyp := rval.Type()
	// val must be a pointer, otherwise we refuse to decode
	if rtyp.Kind() != reflect.Ptr {
		return newDecodeError("can only decode into pointer target", rtyp)
	}
	if rval.IsNil() {
		return newDecodeError("cannot output to pointer of nil", rtyp)
	}
	sszUtils, err := cachedSSZUtils(rval.Elem().Type())
	if err != nil {
		return newDecodeError(fmt.Sprint(err), rval.Elem().Type())
	}
	if _, err = sszUtils.decoder(input, rval.Elem()); err != nil {
		return newDecodeError(fmt.Sprint(err), rval.Elem().Type())
	}
	return nil
}

func makeDecoder(typ reflect.Type) (dec decoder, err error) {
	kind := typ.Kind()
	switch {
	case kind == reflect.Bool:
		return decodeBool, nil
	case kind == reflect.Uint8:
		return decodeUint8, nil
	case kind == reflect.Uint16:
		return decodeUint16, nil
	case kind == reflect.Uint32:
		return decodeUint32, nil
	case kind == reflect.Int32:
		return decodeUint32, nil
	case kind == reflect.Uint64:
		return decodeUint64, nil
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		return decodeBytes, nil
	case kind == reflect.Array && typ.Elem().Kind() == reflect.Uint8:
		return decodeBytes, nil
	case kind == reflect.Slice && isBasicType(typ.Elem().Kind()):
		return makeBasicSliceDecoder(typ)
	case kind == reflect.Slice && !isBasicType(typ.Elem().Kind()):
		return makeCompositeSliceDecoder(typ)
	case kind == reflect.Array:
		return makeArrayDecoder(typ)
	case kind == reflect.Struct:
		return makeStructDecoder(typ)
	case kind == reflect.Ptr:
		return makePtrDecoder(typ)
	default:
		return nil, fmt.Errorf("type %v is not deserializable", typ)
	}
}

func decodeBool(input []byte, val reflect.Value) (int, error) {
	v := uint8(input[0])
	if v == 0 {
		val.SetBool(false)
	} else if v == 1 {
		val.SetBool(true)
	} else {
		return 0, fmt.Errorf("expect 0 or 1 for decoding bool but got %d", v)
	}
	return 1, nil
}

func decodeUint8(input []byte, val reflect.Value) (int, error) {
	val.SetUint(uint64(input[0]))
	return 1, nil
}

func decodeUint16(input []byte, val reflect.Value) (int, error) {
	buf := make([]byte, 2)
	copy(buf, input)
	val.SetUint(uint64(binary.LittleEndian.Uint16(buf)))
	return 2, nil
}

func decodeUint32(input []byte, val reflect.Value) (int, error) {
	buf := make([]byte, 4)
	copy(buf, input)
	val.SetUint(uint64(binary.LittleEndian.Uint32(buf)))
	return 4, nil
}

func decodeUint64(input []byte, val reflect.Value) (int, error) {
	buf := make([]byte, 8)
	copy(buf, input)
	val.SetUint(binary.LittleEndian.Uint64(buf))
	return 8, nil
}

func decodeBytes(input []byte, val reflect.Value) (int, error) {
	val.SetBytes(input)
	return len(input), nil
}

func makeBasicSliceDecoder(typ reflect.Type) (decoder, error) {
	elemType := typ.Elem()
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(elemType)
	if err != nil {
		return nil, err
	}
	decoder := func(r io.Reader, val reflect.Value) (int, error) {
		bufferedReader := bufio.NewReader(r)
		res, err := bufferedReader.ReadBytes('\n')
		if err != nil {
			return 0, fmt.Errorf("hi hi: %v", err)
		}
		fmt.Println(res)
        size := bufferedReader.Size() / basicElementSize(typ.Elem().Kind())
		for i, decodeSize := 0, uint64(0); i < size; i++ {
			// Grow slice's capacity if necessary
			if i >= val.Cap() {
				fmt.Println("Growing")
				newCap := val.Cap() * 2
				// Skip initial small growth
				if newCap < 4 {
					newCap = 4
				}
				newVal := reflect.MakeSlice(val.Type(), val.Len(), newCap)
				reflect.Copy(newVal, val)
				val.Set(newVal)
			}

			// Add place holder for new element
			if i >= val.Len() {
				val.SetLen(i + 1)
			}

			fmt.Println(val.Index(i))
			// Decode and write into the new element
			elemDecodeSize, err := elemSSZUtils.decoder(r, val.Index(i))
			if err != nil {
				return 0, fmt.Errorf("failed to decode element of slice: %v", err)
			}
			fmt.Println(val)
			decodeSize += uint64(elemDecodeSize)
		}
		return int(lengthBytes + size), nil
	}
	return decoder, nil
}

func makeCompositeSliceDecoder(typ reflect.Type) (decoder, error) {
	return nil, nil
}

func makeArrayDecoder(typ reflect.Type) (decoder, error) {
	elemType := typ.Elem()
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(elemType)
	if err != nil {
		return nil, err
	}
	decoder := func(r io.Reader, val reflect.Value) (int, error) {
		size := uint64(val.Len())
		i, decodeSize := 0, uint64(0)
		for ; i < val.Len(); i++ {
			elemDecodeSize, err := elemSSZUtils.decoder(r, val.Index(i))
			if err != nil {
				return 0, fmt.Errorf("failed to decode element of slice: %v", err)
			}
			decodeSize += uint64(elemDecodeSize)
		}
		if decodeSize < size {
			return 0, errors.New("input is too long")
		}
		return int(decodeSize), nil
	}
	return decoder, nil
}

func makeStructDecoder(typ reflect.Type) (decoder, error) {
	fields, err := structFields(typ)
	if err != nil {
		return nil, err
	}
	decoder := func(r io.Reader, val reflect.Value) (int, error) {
		buf := []byte{}
		if _, err := r.Read(buf); err != nil {
			return 0, err
		}
		size := binary.LittleEndian.Uint64(buf)

		if size == 0 {
			return lengthBytes, nil
		}

		i, decodeSize := 0, uint64(0)
		for ; i < len(fields) && decodeSize < size; i++ {
			f := fields[i]
			fieldDecodeSize, err := f.sszUtils.decoder(r, val.Field(f.index))
			if err != nil {
				return 0, fmt.Errorf("failed to decode field of slice: %v", err)
			}
			decodeSize += uint64(fieldDecodeSize)
		}
		if i < len(fields) {
			return 0, errors.New("input is too short")
		}
		if decodeSize < size {
			return 0, errors.New("input is too long")
		}
		return int(lengthBytes + size), nil
	}
	return decoder, nil
}

func makePtrDecoder(typ reflect.Type) (decoder, error) {
	elemType := typ.Elem()
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(elemType)
	if err != nil {
		return nil, err
	}

	// After considered the use case in Prysm, we've decided that:
	// - We assume we will only encode/decode pointer of array, slice or struct.
	// - The encoding for nil pointer shall be 0x00000000.

	decoder := func(r io.Reader, val reflect.Value) (int, error) {
		newVal := reflect.New(elemType)
		elemDecodeSize, err := elemSSZUtils.decoder(r, newVal.Elem())
		if err != nil {
			return 0, fmt.Errorf("failed to decode to object pointed by pointer: %v", err)
		}
		if elemDecodeSize > lengthBytes {
			val.Set(newVal)
		} // Else we leave val to its default value which is nil.
		return elemDecodeSize, nil
	}
	return decoder, nil
}
