package ssz

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

type encbuf struct {
	str []byte
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

// Encode encodes val and output the result into w.
func Encode(w io.Writer, val interface{}) error {
	eb := &encbuf{}
	// Need to preallocate
	if err := eb.encode(val); err != nil {
		return err
	}
	return eb.toWriter(w)
}

func (w *encbuf) encode(val interface{}) error {
	if val == nil {
		return newEncodeError("untyped nil is not supported", nil)
	}
	rval := reflect.ValueOf(val)
	// We preallocate a buffer-size depending on the value's size:
	valueSize := 0
	w.str = make([]byte, valueSize)
	sszUtils, err := cachedSSZUtils(rval.Type())
	if err != nil {
		return newEncodeError(fmt.Sprint(err), rval.Type())
	}
	if _, err = sszUtils.encoder(rval, w, 0 /* start offset */); err != nil {
		return newEncodeError(fmt.Sprint(err), rval.Type())
	}
	return nil
}

func (w *encbuf) toWriter(out io.Writer) error {
	_, err := out.Write(w.str)
	return err
}

func makeEncoder(typ reflect.Type) (encoder, error) {
	kind := typ.Kind()
	switch {
	case kind == reflect.Bool:
		return encodeBool, nil
	case kind == reflect.Uint8:
		return encodeUint8, nil
	case kind == reflect.Uint16:
		return encodeUint16, nil
	case kind == reflect.Uint32:
		return encodeUint32, nil
	case kind == reflect.Uint64:
		return encodeUint64, nil
	case (kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8) ||
		(kind == reflect.Array && typ.Elem().Kind() == reflect.Uint8):
		return encodeBytes, nil
	case kind == reflect.Slice || kind == reflect.Array:
		return makeSliceEncoder(typ)
	//case kind == reflect.Struct:
	//	return makeStructEncoder(typ)
	//case kind == reflect.Ptr:
	//	return makePtrEncoder(typ)
	default:
		return nil, fmt.Errorf("type %v is not serializable", typ)
	}
}

func encodeBool(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
	if val.Bool() {
		w.str = append(w.str, uint8(1))
	} else {
		w.str = append(w.str, uint8(0))
	}
	return startOffset + 1, nil
}

func encodeUint8(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
	v := val.Uint()
	w.str = append(w.str, uint8(v))
	return startOffset + 1, nil
}

func encodeUint16(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
	v := val.Uint()
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(v))
	w.str = append(w.str, b...)
	return startOffset + 2, nil
}

func encodeUint32(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
	v := val.Uint()
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(v))
	w.str = append(w.str, b...)
	return startOffset + 4, nil
}

func encodeUint64(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
	v := val.Uint()
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	w.str = append(w.str, b...)
	return startOffset + 8, nil
}

func encodeBytes(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
	w.str = append(w.str, val.Bytes()...)
	return startOffset + uint64(val.Len()), nil
}

func makeSliceEncoder(typ reflect.Type) (encoder, error) {
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, fmt.Errorf("failed to get ssz utils: %v", err)
	}

	encoder := func(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
		index := startOffset
		var err error
		if isBasicType(typ.Elem().Kind()) || typ.Elem().Kind() == reflect.Array {
			for i := 0; i < val.Len(); i++ {
				index, err = elemSSZUtils.encoder(val.Index(i), w, index)
				if err != nil {
					return 0, err
				}
			}
		} else {
			fixedIndex := index
			currentOffsetIndex := startOffset + uint64(val.Len()*BytesPerLengthOffset)
			nextOffsetIndex := currentOffsetIndex
			// If the elements are variable size, we need to include offset indices
			// in the serialized output list.
			for i := 0; i < val.Len(); i++ {
				nextOffsetIndex, err = elemSSZUtils.encoder(val.Index(i), w, currentOffsetIndex)
				if err != nil {
					return 0, err
				}
				// Write the offset.
				skipIndices := make([]byte, fixedIndex)
				w.str = append(w.str, skipIndices...)
				offsetBuf := make([]byte, BytesPerLengthOffset)
				binary.LittleEndian.PutUint32(offsetBuf, uint32(currentOffsetIndex-startOffset))
				w.str = append(w.str, offsetBuf...)
				currentOffsetIndex = nextOffsetIndex
				fixedIndex += uint64(BytesPerLengthOffset)
			}
			index = currentOffsetIndex
		}
		return index, nil
	}
	return encoder, nil
}

//func makeStructEncoder(typ reflect.Type) (encoder, error) {
//	fields, err := structFields(typ)
//	if err != nil {
//		return nil, err
//	}
//	encoder := func(val reflect.Value, w *encbuf) error {
//		fixedParts := [][]byte{}
//		variableParts := [][]byte{}
//		for _, f := range fields {
//			item := val.Field(f.index)
//			// Determine the fixed parts of the element.
//			if isBasicType(item.Kind()) || item.Kind() == reflect.Array {
//				elemBuf := &encbuf{}
//				if err := f.sszUtils.encoder(item, w); err != nil {
//					return fmt.Errorf("failed to encode field of struct: %v", err)
//				}
//				fixedParts = append(fixedParts, elemBuf.str)
//			} else {
//				elemBuf := &encbuf{}
//				if err := f.sszUtils.encoder(item, w); err != nil {
//					return fmt.Errorf("failed to encode field of struct: %v", err)
//				}
//				variableParts = append(variableParts, elemBuf.str)
//				fixedParts = append(fixedParts, []byte{})
//			}
//		}
//		serializedStruct, err := serializeFromParts(fixedParts, variableParts, len(fields))
//		if err != nil {
//			return err
//		}
//		w.str = append(w.str, serializedStruct...)
//		return nil
//	}
//	return encoder, nil
//}

//func makePtrEncoder(typ reflect.Type) (encoder, error) {
//	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
//	if err != nil {
//		return nil, err
//	}
//	encoder := func(val reflect.Value, w *encbuf) error {
//		// Nil encodes to []byte{}.
//		if val.IsNil() {
//			w.str = append(w.str, []byte{}...)
//			return nil
//		}
//		return elemSSZUtils.encoder(val.Elem(), w)
//	}
//
//	return encoder, nil
//}

//func serializeFromParts(fixedParts [][]byte, variableParts [][]byte, numElements int) ([]byte, error) {
//	fixedLengths := []int{}
//	variableLengths := []int{}
//	for _, item := range fixedParts {
//		if !bytes.Equal(item, []byte{}) {
//			fixedLengths = append(fixedLengths, len(item))
//		} else {
//			fixedLengths = append(fixedLengths, BytesPerLengthOffset)
//		}
//	}
//	for _, item := range variableParts {
//		if len(item) == 0 {
//			continue
//		}
//		variableLengths = append(variableLengths, len(item))
//	}
//	sum := 0
//	for _, item := range append(fixedLengths, variableLengths...) {
//		sum += item
//	}
//	if sum >= MaxByteOffset {
//		return nil, fmt.Errorf(
//			"expected sum(fixed_length + variable_length) < MaxByteOffset, received %d >= %d",
//			sum,
//			MaxByteOffset,
//		)
//	}
//	variableOffsets := [][]byte{}
//	upperBound := numElements
//	if len(variableLengths) < upperBound {
//		upperBound = len(variableLengths)
//	}
//	for i := 0; i < upperBound; i++ {
//		sum = 0
//		for _, item := range append(fixedLengths, variableLengths[:i]...) {
//			sum += item
//		}
//		b := make([]byte, 8)
//		binary.LittleEndian.PutUint64(b, uint64(sum))
//		variableOffsets = append(variableOffsets, b)
//	}
//	offsetParts := [][]byte{}
//	for idx, item := range fixedParts {
//		if !bytes.Equal(item, []byte{}) {
//			offsetParts = append(offsetParts, item)
//		} else {
//			if idx < len(variableOffsets) {
//				offsetParts = append(offsetParts, variableOffsets[idx])
//			}
//		}
//	}
//	fixedParts = offsetParts
//	concat := append(fixedParts, variableParts...)
//	finalSerialization := []byte{}
//	// We flatten the final serialized slice.
//	for _, item := range concat {
//		finalSerialization = append(finalSerialization, item...)
//	}
//	return finalSerialization, nil
//}
