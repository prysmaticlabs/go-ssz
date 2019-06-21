package ssz

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

// marshalError is what gets reported to the marshaler user in error case.
type marshalError struct {
	msg string
	typ reflect.Type
}

func newMarshalError(msg string, typ reflect.Type) *marshalError {
	return &marshalError{msg, typ}
}

func (err *marshalError) Error() string {
	return fmt.Sprintf("marshal error: %s for input type %v", err.msg, err.typ)
}

// Marshal a value and output the result into a byte slice.
func Marshal(val interface{}) ([]byte, error) {
	if val == nil {
		return nil, newMarshalError("untyped nil is not supported", nil)
	}
	rval := reflect.ValueOf(val)

	// We pre-allocate a buffer-size depending on the value's size.
	buf := make([]byte, determineSize(rval))
	fmt.Println(len(buf))
	sszUtils, err := cachedSSZUtils(rval.Type())
	if err != nil {
		return nil, newMarshalError(fmt.Sprint(err), rval.Type())
	}
	if _, err = sszUtils.marshaler(rval, buf, 0 /* start offset */); err != nil {
		return nil, newMarshalError(fmt.Sprint(err), rval.Type())
	}
	return buf, nil
}

func makeMarshaler(typ reflect.Type) (marshaler, error) {
	kind := typ.Kind()
	switch {
	case kind == reflect.Bool:
		return marshalBool, nil
	case kind == reflect.Uint8:
		return marshalUint8, nil
	case kind == reflect.Uint16:
		return marshalUint16, nil
	case kind == reflect.Uint32:
		return marshalUint32, nil
	case kind == reflect.Uint64:
		return marshalUint64, nil
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		return marshalByteSlice, nil
	case kind == reflect.Array && typ.Elem().Kind() == reflect.Uint8:
		return marshalByteArray, nil
	case kind == reflect.Slice && isBasicTypeArray(typ.Elem(), typ.Elem().Kind()):
		return makeBasicSliceMarshaler(typ)
	case kind == reflect.Slice && isBasicType(typ.Elem().Kind()):
		return makeBasicSliceMarshaler(typ)
	case kind == reflect.Slice && !isVariableSizeType(typ.Elem()):
		return makeBasicSliceMarshaler(typ)
	case kind == reflect.Slice || kind == reflect.Array:
		return makeCompositeSliceMarshaler(typ)
	case kind == reflect.Struct:
		return makeStructMarshaler(typ)
	case kind == reflect.Ptr:
		return makePtrMarshaler(typ)
	default:
		return nil, fmt.Errorf("type %v is not serializable", typ)
	}
}

func marshalBool(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	if val.Bool() {
		buf[startOffset] = uint8(1)
	} else {
		buf[startOffset] = uint8(0)
	}
	return startOffset + 1, nil
}

func marshalUint8(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	v := val.Uint()
	buf[startOffset] = uint8(v)
	return startOffset + 1, nil
}

func marshalUint16(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	v := val.Uint()
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(v))
	copy(buf[startOffset:startOffset+2], b)
	return startOffset + 2, nil
}

func marshalUint32(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	v := val.Uint()
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(v))
	copy(buf[startOffset:startOffset+4], b)
	return startOffset + 4, nil
}

func marshalUint64(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	v := val.Uint()
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	copy(buf[startOffset:startOffset+8], b)
	return startOffset + 8, nil
}

func marshalByteSlice(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	slice := val.Slice(0, val.Len()).Interface().([]byte)
	copy(buf[startOffset:startOffset+uint64(len(slice))], slice)
	return startOffset + uint64(val.Len()), nil
}

func marshalByteArray(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	rawBytes := make([]byte, val.Len())
	for i := 0; i < val.Len(); i++ {
		rawBytes[i] = uint8(val.Index(i).Uint())
	}
	copy(buf[startOffset:startOffset+uint64(len(rawBytes))], rawBytes)
	return startOffset + uint64(len(rawBytes)), nil
}

func makeBasicSliceMarshaler(typ reflect.Type) (marshaler, error) {
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, fmt.Errorf("failed to get ssz utils: %v", err)
	}

	marshaler := func(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
		index := startOffset
		var err error
		for i := 0; i < val.Len(); i++ {
			index, err = elemSSZUtils.marshaler(val.Index(i), buf, index)
			if err != nil {
				return 0, err
			}
		}
		return index, nil
	}
	return marshaler, nil
}

func makeCompositeSliceMarshaler(typ reflect.Type) (marshaler, error) {
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, fmt.Errorf("failed to get ssz utils: %v", err)
	}

	marshaler := func(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
		index := startOffset
		var err error
		if !isVariableSizeType(typ) {
			for i := 0; i < val.Len(); i++ {
				// If each element is not variable size, we simply encode sequentially and write
				// into the buffer at the last index we wrote at.
				index, err = elemSSZUtils.marshaler(val.Index(i), buf, index)
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
				nextOffsetIndex, err = elemSSZUtils.marshaler(val.Index(i), buf, currentOffsetIndex)
				if err != nil {
					return 0, err
				}
				// Write the offset.
				offsetBuf := make([]byte, BytesPerLengthOffset)
				binary.LittleEndian.PutUint32(offsetBuf, uint32(currentOffsetIndex-startOffset))
				copy(buf[fixedIndex:fixedIndex+uint64(BytesPerLengthOffset)], offsetBuf)

				// We increase the offset indices accordingly.
				currentOffsetIndex = nextOffsetIndex
				fixedIndex += uint64(BytesPerLengthOffset)
			}
			index = currentOffsetIndex
		}
		return index, nil
	}
	return marshaler, nil
}

func makeStructMarshaler(typ reflect.Type) (marshaler, error) {
	fields, err := marshalerStructFields(typ)
	if err != nil {
		return nil, err
	}
	marshaler := func(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
		fixedIndex := startOffset
		fixedLength := uint64(0)
		// For every field, we add up the total length of the items depending if they
		// are variable or fixed-size fields.
		for _, f := range fields {
			if isVariableSizeType(f.typ) {
				fixedLength += uint64(BytesPerLengthOffset)
			} else {
				fixedLength += determineFixedSize(val.Field(f.index), f.typ)
			}
		}
		currentOffsetIndex := startOffset + fixedLength
		nextOffsetIndex := currentOffsetIndex
		var err error
		for i, f := range fields {
			if !isVariableSizeType(f.typ) {
				fixedIndex, err = f.sszUtils.marshaler(val.Field(i), buf, fixedIndex)
				if err != nil {
					return 0, err
				}
			} else {
				nextOffsetIndex, err = f.sszUtils.marshaler(val.Field(f.index), buf, currentOffsetIndex)
				if err != nil {
					return 0, err
				}
				// Write the offset.
				offsetBuf := make([]byte, BytesPerLengthOffset)
				binary.LittleEndian.PutUint32(offsetBuf, uint32(currentOffsetIndex-startOffset))
				copy(buf[fixedIndex:fixedIndex+uint64(BytesPerLengthOffset)], offsetBuf)

				// We increase the offset indices accordingly.
				currentOffsetIndex = nextOffsetIndex
				fixedIndex += uint64(BytesPerLengthOffset)
			}
		}
		return currentOffsetIndex, nil
	}
	return marshaler, nil
}

func makePtrMarshaler(typ reflect.Type) (marshaler, error) {
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, err
	}
	marshaler := func(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
		// Nil encodes to []byte{}.
		if val.IsNil() {
			buf[startOffset] = 0
			return 0, nil
		}
		return elemSSZUtils.marshaler(val.Elem(), buf, startOffset)
	}

	return marshaler, nil
}
