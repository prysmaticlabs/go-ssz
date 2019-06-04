package ssz

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

const lengthBytes = 4

// Encode encodes val and output the result into w.
func Encode(w io.Writer, val interface{}) error {
	eb := &encbuf{}
	if err := eb.encode(val); err != nil {
		return err
	}
	return eb.toWriter(w)
}

// EncodeSize returns the target encoding size without doing the actual encoding.
// This is an optional pass. You don't need to call this before the encoding unless you
// want to know the output size first.
func EncodeSize(val interface{}) (uint32, error) {
	return encodeSize(val)
}

type encbuf struct {
	str []byte
}

func (w *encbuf) encode(val interface{}) error {
	if val == nil {
		return newEncodeError("untyped nil is not supported", nil)
	}
	rval := reflect.ValueOf(val)
	sszUtils, err := cachedSSZUtils(rval.Type())
	if err != nil {
		return newEncodeError(fmt.Sprint(err), rval.Type())
	}
	if err = sszUtils.encoder(rval, w); err != nil {
		return newEncodeError(fmt.Sprint(err), rval.Type())
	}
	return nil
}

func encodeSize(val interface{}) (uint32, error) {
	if val == nil {
		return 0, newEncodeError("untyped nil is not supported", nil)
	}
	rval := reflect.ValueOf(val)
	sszUtils, err := cachedSSZUtils(rval.Type())
	if err != nil {
		return 0, newEncodeError(fmt.Sprint(err), rval.Type())
	}
	var size uint32
	if size, err = sszUtils.encodeSizer(rval); err != nil {
		return 0, newEncodeError(fmt.Sprint(err), rval.Type())
	}
	return size, nil

}

func (w *encbuf) toWriter(out io.Writer) error {
	_, err := out.Write(w.str)
	return err
}

func makeEncoder(typ reflect.Type) (encoder, encodeSizer, error) {
	kind := typ.Kind()
	switch {
	case kind == reflect.Bool:
		return encodeBool, func(reflect.Value) (uint32, error) { return 1, nil }, nil
	case kind == reflect.Uint8:
		return encodeUint8, func(reflect.Value) (uint32, error) { return 1, nil }, nil
	case kind == reflect.Uint16:
		return encodeUint16, func(reflect.Value) (uint32, error) { return 2, nil }, nil
	case kind == reflect.Uint32:
		return encodeUint32, func(reflect.Value) (uint32, error) { return 4, nil }, nil
	case kind == reflect.Uint64:
		return encodeUint64, func(reflect.Value) (uint32, error) { return 8, nil }, nil
	case kind == reflect.Slice:
		return makeSliceEncoder(typ)
	case kind == reflect.Array:
		return makeSliceEncoder(typ)
	case kind == reflect.Struct:
		return makeStructEncoder(typ)
	case kind == reflect.Ptr:
		return makePtrEncoder(typ)
	default:
		return nil, nil, fmt.Errorf("type %v is not serializable", typ)
	}
}

func encodeBool(val reflect.Value, w *encbuf) error {
	if val.Bool() {
		w.str = append(w.str, uint8(1))
	} else {
		w.str = append(w.str, uint8(0))
	}
	return nil
}

func encodeUint8(val reflect.Value, w *encbuf) error {
	v := val.Uint()
	w.str = append(w.str, uint8(v))
	return nil
}

func encodeUint16(val reflect.Value, w *encbuf) error {
	v := val.Uint()
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(v))
	w.str = append(w.str, b...)
	return nil
}

func encodeUint32(val reflect.Value, w *encbuf) error {
	v := val.Uint()
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(v))
	w.str = append(w.str, b...)
	return nil
}

func encodeUint64(val reflect.Value, w *encbuf) error {
	v := val.Uint()
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	w.str = append(w.str, b...)
	return nil
}

func makeSliceEncoder(typ reflect.Type) (encoder, encodeSizer, error) {
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get ssz utils: %v", err)
	}
	encoder := func(val reflect.Value, w *encbuf) error {
		fixedParts := [][]byte{}
		variableParts := [][]byte{}
		for i := 0; i < val.Len(); i++ {
            // Determine the fixed parts of the element.
			if isBasicType(typ.Elem().Kind()) || typ.Elem().Kind() == reflect.Array {
				elemBuf := &encbuf{}
				if err := elemSSZUtils.encoder(val.Index(i), elemBuf); err != nil {
					return fmt.Errorf("failed to encode element of slice/array: %v", err)
				}
				fixedParts = append(fixedParts, elemBuf.str)
			} else {
				elemBuf := &encbuf{}
				if err := elemSSZUtils.encoder(val.Index(i), elemBuf); err != nil {
					return fmt.Errorf("failed to encode element of slice/array: %v", err)
				}
				variableParts = append(variableParts, elemBuf.str)
				fixedParts = append(fixedParts, []byte{})
			}
		}
		fixedLengths := []int{}
		variableLengths := []int{}
		for _, item := range fixedParts {
			if !bytes.Equal(item, []byte{}) {
				fixedLengths = append(fixedLengths, len(item))
			} else {
				fixedLengths = append(fixedLengths, BytesPerLengthOffset)
			}
		}
		for _, item := range variableParts {
			variableLengths = append(variableLengths, len(item))
		}
        sum := 0
        for _, item := range append(fixedLengths, variableLengths...) {
        	sum += item
		}
        if sum >= 1<<uint64(BytesPerLengthOffset*BitsPerByte) {
        	return fmt.Errorf(
        		"expected sum(fixed_length + variable_length) < 2**(BytesPerLengthOffset*BitsPerByte), received %d >= %d",
        		sum,
        		1<<uint64(BytesPerLengthOffset*BitsPerByte),
        	)
		}
        variableOffsets := [][]byte{}
        for i := 0; i < val.Len(); i++ {
        	sum = 0
        	for _, item := range append(fixedLengths, variableLengths[:i]...) {
        		sum += item
			}
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(sum))
			variableOffsets = append(variableOffsets, b)
		}
        offsetParts := [][]byte{}
        for idx, item := range fixedParts {
        	if !bytes.Equal(item, []byte{}) {
        		offsetParts = append(offsetParts, item)
			} else {
				offsetParts = append(offsetParts, variableOffsets[idx])
			}
		}
        fixedParts = offsetParts
		concat := append(fixedParts, variableParts...)
       	finalSerialization := []byte{}
       	// We flatten the final serialized slice.
       	for _, item := range concat {
       		finalSerialization = append(finalSerialization, item...)
		}
		w.str = append(w.str, finalSerialization...)
		return nil
	}
	encodeSizer := func(val reflect.Value) (uint32, error) {
		return 0, nil
	}
	return encoder, encodeSizer, nil
}

func makeStructEncoder(typ reflect.Type) (encoder, encodeSizer, error) {
	fields, err := structFields(typ)
	if err != nil {
		return nil, nil, err
	}
	encoder := func(val reflect.Value, w *encbuf) error {
		origBufSize := len(w.str)
		totalSizeEnc := make([]byte, lengthBytes)
		w.str = append(w.str, totalSizeEnc...)
		for _, f := range fields {
			if err := f.sszUtils.encoder(val.Field(f.index), w); err != nil {
				return fmt.Errorf("failed to encode field of struct: %v", err)
			}
		}
		totalSize := len(w.str) - lengthBytes - origBufSize
		if totalSize >= 2<<32 {
			return errors.New("struct oversize")
		}
		binary.LittleEndian.PutUint32(totalSizeEnc, uint32(totalSize))
		copy(w.str[origBufSize:origBufSize+lengthBytes], totalSizeEnc)
		return nil
	}
	encodeSizer := func(val reflect.Value) (uint32, error) {
		totalSize := uint32(0)
		for _, f := range fields {
			fieldSize, err := f.sszUtils.encodeSizer(val.Field(f.index))
			if err != nil {
				return 0, fmt.Errorf("failed to get encode size for field of struct: %v", err)
			}
			totalSize += fieldSize
		}
		return lengthBytes + totalSize, nil
	}
	return encoder, encodeSizer, nil
}

func makePtrEncoder(typ reflect.Type) (encoder, encodeSizer, error) {
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, nil, err
	}

	// After considered the use case in Prysm, we've decided that:
	// - We assume we will only encode/decode pointer of array, slice or struct.
	// - The encoding for nil pointer shall be 0x00000000.
	encoder := func(val reflect.Value, w *encbuf) error {
		if val.IsNil() {
			totalSizeEnc := make([]byte, lengthBytes)
			w.str = append(w.str, totalSizeEnc...)
			return nil
		}
		return elemSSZUtils.encoder(val.Elem(), w)
	}

	encodeSizer := func(val reflect.Value) (uint32, error) {
		if val.IsNil() {
			return lengthBytes, nil
		}
		return elemSSZUtils.encodeSizer(val.Elem())
	}

	return encoder, encodeSizer, nil
}

// encodeError is what gets reported to the encoder user in error case.
type encodeError struct {
	msg string
	typ reflect.Type
}

func newEncodeError(msg string, typ reflect.Type) *encodeError {
	return &encodeError{msg, typ}
}

func (err *encodeError) Error() string {
	return fmt.Sprintf("encode error: %s for input type %v", err.msg, err.typ)
}

func isBasicType(kind reflect.Kind) bool {
	return kind == reflect.Bool ||
		kind == reflect.Uint8 ||
		kind == reflect.Uint16 ||
		kind == reflect.Uint32 ||
		kind == reflect.Uint64
}
