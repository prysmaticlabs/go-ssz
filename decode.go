package ssz

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
)

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
	if err = sszUtils.decoder(input, rval.Elem(), 0); err != nil {
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
		return makeByteSliceDecoder()
	case kind == reflect.Slice && isBasicType(typ.Elem().Kind()):
		return makeBasicSliceDecoder(typ)
	case kind == reflect.Slice && !isBasicType(typ.Elem().Kind()):
		return makeCompositeSliceDecoder(typ)
	case kind == reflect.Array:
		return makeArrayDecoder(typ)
	//case kind == reflect.Struct:
	//	return makeStructDecoder(typ)
	//case kind == reflect.Ptr:
	//	return makePtrDecoder(typ)
	default:
		return nil, fmt.Errorf("type %v is not deserializable", typ)
	}
}

func decodeBool(input []byte, val reflect.Value, startOffset uint64) error {
	v := uint8(input[startOffset])
	if v == 0 {
		val.SetBool(false)
	} else if v == 1 {
		val.SetBool(true)
	} else {
		return fmt.Errorf("expect 0 or 1 for decoding bool but got %d", v)
	}
	return nil
}

func decodeUint8(input []byte, val reflect.Value, startOffset uint64) error {
	val.SetUint(uint64(input[startOffset]))
	return nil
}

func decodeUint16(input []byte, val reflect.Value, startOffset uint64) error {
	offset := startOffset + 2
	buf := make([]byte, 2)
	copy(buf, input[startOffset:offset])
	val.SetUint(uint64(binary.LittleEndian.Uint16(buf)))
	return nil
}

func decodeUint32(input []byte, val reflect.Value, startOffset uint64) error {
	offset := startOffset + 4
	buf := make([]byte, 4)
	copy(buf, input[startOffset:offset])
	val.SetUint(uint64(binary.LittleEndian.Uint32(buf)))
	return nil
}

func decodeUint64(input []byte, val reflect.Value, startOffset uint64) error {
	offset := startOffset + 8
	buf := make([]byte, 8)
	copy(buf, input[startOffset:offset])
	fmt.Println(buf)
	val.SetUint(binary.LittleEndian.Uint64(buf))
	return nil
}

func makeByteSliceDecoder() (decoder, error) {
	decoder := func(input []byte, val reflect.Value, startOffset uint64) error {
		val.SetBytes(input[startOffset:uint64(len(input))])
		return nil
	}
	return decoder, nil
}

func makeBasicSliceDecoder(typ reflect.Type) (decoder, error) {
	elemType := typ.Elem()
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(elemType)
	if err != nil {
		return nil, err
	}
	decoder := func(input []byte, val reflect.Value, startOffset uint64) error {
		index := startOffset
		elementSize := basicTypeSize(typ.Elem())
		endOffset := uint64(len(input)) / elementSize
		if startOffset == endOffset {
			return nil
		}
		newVal := reflect.MakeSlice(val.Type(), int(endOffset), int(endOffset))
		reflect.Copy(newVal, val)
		val.Set(newVal)
		i := uint64(0)
		var nextIndex uint64
		for i < endOffset {
			nextIndex = index + elementSize
			if err := elemSSZUtils.decoder(input, val.Index(int(i)), index); err != nil {
				return fmt.Errorf("failed to decode element of slice: %v", err)
			}
			index = nextIndex
			i++
		}
		return nil
	}
	return decoder, nil
}

func makeCompositeSliceDecoder(typ reflect.Type) (decoder, error) {
	elemType := typ.Elem()
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(elemType)
	if err != nil {
		return nil, err
	}
	decoder := func(input []byte, val reflect.Value, startOffset uint64) error {
		newVal := reflect.MakeSlice(val.Type(), 1, 1)
		reflect.Copy(newVal, val)
		val.Set(newVal)
		endOffset := uint64(len(input))

		currentIndex := startOffset
		nextIndex := currentIndex
		offsetVal := input[startOffset : startOffset+uint64(BytesPerLengthOffset)]
		firstOffset := startOffset + uint64(binary.LittleEndian.Uint32(offsetVal))
		currentOffset := firstOffset
		nextOffset := currentOffset
		i := 0
		for currentIndex < firstOffset {
			if currentOffset > endOffset {
				return errors.New("offset out of bounds")
			}
			nextIndex = currentIndex + uint64(BytesPerLengthOffset)
			if nextIndex == firstOffset {
				nextOffset = endOffset
			} else {
				nextOffsetVal := input[nextIndex : nextIndex+uint64(BytesPerLengthOffset)]
				nextOffset = startOffset + uint64(binary.LittleEndian.Uint32(nextOffsetVal))
			}
			if currentOffset > nextOffset {
				return errors.New("offsets must be increasing")
			}
			// We grow the slice's size to accommodate a new element being decoded.
			newVal := reflect.MakeSlice(val.Type(), i+1, i+1)
			reflect.Copy(newVal, val)
			val.Set(newVal)
			if err := elemSSZUtils.decoder(input[currentOffset:nextOffset], val.Index(i), 0); err != nil {
				return fmt.Errorf("failed to decode element of slice: %v", err)
			}
			i++
			currentIndex = nextIndex
			currentOffset = nextOffset
		}
		return nil
	}
	return decoder, nil
}

func makeArrayDecoder(typ reflect.Type) (decoder, error) {
	elemType := typ.Elem()
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(elemType)
	if err != nil {
		return nil, err
	}
	decoder := func(input []byte, val reflect.Value, startOffset uint64) error {
		i := 0
		index := startOffset
		elemSize := basicTypeSize(typ.Elem())
		size := val.Len()
		var nextIndex uint64
		for i < size {
			nextIndex = index + elemSize
			if err := elemSSZUtils.decoder(input, val.Index(i), index); err != nil {
				return fmt.Errorf("failed to decode element of array: %v", err)
			}
			index = nextIndex
			i++
		}
		return nil
	}
	return decoder, nil
}

//func makeStructDecoder(typ reflect.Type) (decoder, error) {
//	fields, err := structFields(typ)
//	if err != nil {
//		return nil, err
//	}
//	decoder := func(input []byte, val reflect.Value) (int, error) {
//		size := len(input)
//
//		if size == 0 {
//			return 0, nil
//		}
//
//		i := 0
//		offsetIndex := 0
//		for ; i < len(fields); i++ {
//			// Track the offset index verifying if a field is variable-size or fixed-size, and then proceed.
//			f := fields[i]
//			// TODO: Handle is variadic.
//			elemSize := basicElementSize(val.Field(f.index).Type(), val.Field(f.index).Kind())
//			fieldDecodeSize, err := f.sszUtils.decoder(input[offsetIndex:offsetIndex+elemSize], val.Field(f.index))
//			if err != nil {
//				return 0, fmt.Errorf("failed to decode field of slice: %v", err)
//			}
//			offsetIndex += fieldDecodeSize
//		}
//		if i < len(fields) {
//			return 0, errors.New("input is too short")
//		}
//		return offsetIndex, nil
//	}
//	return decoder, nil
//}
//
//func makePtrDecoder(typ reflect.Type) (decoder, error) {
//	elemType := typ.Elem()
//	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(elemType)
//	if err != nil {
//		return nil, err
//	}
//	decoder := func(input []byte, val reflect.Value) (int, error) {
//		newVal := reflect.New(elemType)
//		elemDecodeSize, err := elemSSZUtils.decoder(input, newVal.Elem())
//		if err != nil {
//			return 0, fmt.Errorf("failed to decode to object pointed by pointer: %v", err)
//		}
//		return elemDecodeSize, nil
//	}
//	return decoder, nil
//}
