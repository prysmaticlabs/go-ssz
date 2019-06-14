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

	// We pre-allocate a buffer-size depending on the value's size.
	w.str = make([]byte, determineSize(rval))
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
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		return encodeByteSlice, nil
	case kind == reflect.Array && typ.Elem().Kind() == reflect.Uint8:
		return encodeByteArray, nil
	case kind == reflect.Array:
		return makeBasicSliceEncoder(typ)
	case kind == reflect.Slice && isBasicTypeArray(typ.Elem(), typ.Elem().Kind()):
		return makeBasicSliceEncoder(typ)
	case kind == reflect.Slice && isBasicType(typ.Elem().Kind()):
		return makeBasicSliceEncoder(typ)
	case kind == reflect.Slice:
		return makeCompositeSliceEncoder(typ)
	case kind == reflect.Struct:
		return makeStructEncoder(typ)
	case kind == reflect.Ptr:
		return makePtrEncoder(typ)
	default:
		return nil, fmt.Errorf("type %v is not serializable", typ)
	}
}

func encodeBool(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
	if val.Bool() {
		w.str[startOffset] = uint8(1)
	} else {
		w.str[startOffset] = uint8(0)
	}
	return startOffset + 1, nil
}

func encodeUint8(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
	v := val.Uint()
	w.str[startOffset] = uint8(v)
	return startOffset + 1, nil
}

func encodeUint16(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
	v := val.Uint()
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(v))
	copy(w.str[startOffset:startOffset+2], b)
	return startOffset + 2, nil
}

func encodeUint32(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
	v := val.Uint()
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(v))
	copy(w.str[startOffset:startOffset+4], b)
	return startOffset + 4, nil
}

func encodeUint64(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
	v := val.Uint()
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	copy(w.str[startOffset:startOffset+8], b)
	return startOffset + 8, nil
}

func encodeByteSlice(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
	slice := val.Slice(0, val.Len()).Interface().([]byte)
	copy(w.str[startOffset:startOffset+uint64(len(slice))], slice)
	return startOffset + uint64(val.Len()), nil
}

func encodeByteArray(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
	rawBytes := make([]byte, val.Len())
	for i := 0; i < val.Len(); i++ {
		rawBytes[i] = uint8(val.Index(i).Uint())
	}
	copy(w.str[startOffset:startOffset+uint64(len(rawBytes))], rawBytes)
	return startOffset + uint64(len(rawBytes)), nil
}

func makeBasicSliceEncoder(typ reflect.Type) (encoder, error) {
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, fmt.Errorf("failed to get ssz utils: %v", err)
	}

	encoder := func(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
		index := startOffset
		var err error
		for i := 0; i < val.Len(); i++ {
			index, err = elemSSZUtils.encoder(val.Index(i), w, index)
			if err != nil {
				return 0, err
			}
		}
		return index, nil
	}
	return encoder, nil
}

func makeCompositeSliceEncoder(typ reflect.Type) (encoder, error) {
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, fmt.Errorf("failed to get ssz utils: %v", err)
	}

	encoder := func(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
		index := startOffset
		var err error
		if !isVariableSizeType(val, typ.Elem()) {
			for i := 0; i < val.Len(); i++ {
				// If each element is not variable size, we simply encode sequentially and write
				// into the buffer at the last index we wrote at.
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
				offsetBuf := make([]byte, BytesPerLengthOffset)
				binary.LittleEndian.PutUint32(offsetBuf, uint32(currentOffsetIndex-startOffset))
				copy(w.str[fixedIndex:fixedIndex+uint64(BytesPerLengthOffset)], offsetBuf)

				// We increase the offset indices accordingly.
				currentOffsetIndex = nextOffsetIndex
				fixedIndex += uint64(BytesPerLengthOffset)
			}
			index = currentOffsetIndex
		}
		return index, nil
	}
	return encoder, nil
}

func makeStructEncoder(typ reflect.Type) (encoder, error) {
	fields, err := structFields(typ)
	if err != nil {
		return nil, err
	}
	encoder := func(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
		fixedIndex := startOffset
		fixedLength := uint64(0)
		// For every field, we add up the total length of the items depending if they
		// are variable or fixed-size fields.
		for _, f := range fields {
			item := val.Field(f.index)
			if isVariableSizeType(item, item.Type()) {
				fixedLength += uint64(BytesPerLengthOffset)
			} else {
				fixedLength += determineFixedSize(val.Field(f.index), val.Field(f.index).Type())
			}
		}
		currentOffsetIndex := startOffset + fixedLength
		nextOffsetIndex := currentOffsetIndex
		var err error
		for _, f := range fields {
			item := val.Field(f.index)
			if !isVariableSizeType(item, item.Type()) {
				fixedIndex, err = f.sszUtils.encoder(item, w, fixedIndex)
				if err != nil {
					return 0, err
				}
			} else {
				nextOffsetIndex, err = f.sszUtils.encoder(val.Field(f.index), w, currentOffsetIndex)
				if err != nil {
					return 0, err
				}
				// Write the offset.
				offsetBuf := make([]byte, BytesPerLengthOffset)
				binary.LittleEndian.PutUint32(offsetBuf, uint32(currentOffsetIndex-startOffset))
				copy(w.str[fixedIndex:fixedIndex+uint64(BytesPerLengthOffset)], offsetBuf)

				// We increase the offset indices accordingly.
				currentOffsetIndex = nextOffsetIndex
				fixedIndex += uint64(BytesPerLengthOffset)
			}
		}
		return currentOffsetIndex, nil
	}
	return encoder, nil
}

func makePtrEncoder(typ reflect.Type) (encoder, error) {
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, err
	}
	encoder := func(val reflect.Value, w *encbuf, startOffset uint64) (uint64, error) {
		// Nil encodes to []byte{}.
		if val.IsNil() {
			w.str[startOffset] = 0
			return 0, nil
		}
		return elemSSZUtils.encoder(val.Elem(), w, startOffset)
	}

	return encoder, nil
}
