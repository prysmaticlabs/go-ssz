package ssz

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
)

// unmarshalError is what gets reported to the user in error case.
type unmarshalError struct {
	msg string
	typ reflect.Type
}

func newUnmarshalError(msg string, typ reflect.Type) *unmarshalError {
	return &unmarshalError{msg, typ}
}

func (err *unmarshalError) Error() string {
	return fmt.Sprintf("unmarshal error: %s for output type %v", err.msg, err.typ)
}

// Unmarshal SSZ encoded data and output it into the object pointed by pointer val.
func Unmarshal(input []byte, val interface{}) error {
	if val == nil {
		return newUnmarshalError("cannot unmarshal into nil", nil)
	}
	rval := reflect.ValueOf(val)
	rtyp := rval.Type()
	// val must be a pointer, otherwise we refuse to unmarshal
	if rtyp.Kind() != reflect.Ptr {
		return newUnmarshalError("can only unmarshal into pointer target", rtyp)
	}
	if rval.IsNil() {
		return newUnmarshalError("cannot output to pointer of nil", rtyp)
	}
	sszUtils, err := cachedSSZUtils(rval.Elem().Type())
	if err != nil {
		return newUnmarshalError(fmt.Sprint(err), rval.Elem().Type())
	}
	if _, err = sszUtils.unmarshaler(input, rval.Elem(), 0); err != nil {
		return newUnmarshalError(fmt.Sprint(err), rval.Elem().Type())
	}
	return nil
}

func makeUnmarshaler(typ reflect.Type) (dec unmarshaler, err error) {
	kind := typ.Kind()
	switch {
	case kind == reflect.Bool:
		return unmarshalBool, nil
	case kind == reflect.Uint8:
		return unmarshalUint8, nil
	case kind == reflect.Uint16:
		return unmarshalUint16, nil
	case kind == reflect.Uint32:
		return unmarshalUint32, nil
	case kind == reflect.Int32:
		return unmarshalUint32, nil
	case kind == reflect.Uint64:
		return unmarshalUint64, nil
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		return makeByteSliceUnmarshaler()
	case kind == reflect.Slice && isBasicType(typ.Elem().Kind()):
		return makeBasicSliceUnmarshaler(typ)
	case kind == reflect.Slice && isBasicTypeArray(typ.Elem(), typ.Elem().Kind()):
		return makeBasicSliceUnmarshaler(typ)
	case kind == reflect.Slice && isBasicTypeArray(typ.Elem(), typ.Elem().Kind()):
		return makeBasicSliceUnmarshaler(typ)
	case kind == reflect.Slice:
		return makeCompositeSliceUnmarshaler(typ)
	case kind == reflect.Array && isBasicType(typ.Elem().Kind()):
		return makeBasicArrayUnmarshaler(typ)
	case kind == reflect.Array && isBasicTypeArray(typ.Elem(), typ.Elem().Kind()):
		return makeBasicArrayUnmarshaler(typ)
	case kind == reflect.Array:
		return makeCompositeArrayUnmarshaler(typ)
	case kind == reflect.Struct:
		return makeStructUnmarshaler(typ)
	case kind == reflect.Ptr:
		return makePtrUnmarshaler(typ)
	default:
		return nil, fmt.Errorf("type %v is not deserializable", typ)
	}
}

func unmarshalBool(input []byte, val reflect.Value, startOffset uint64) (uint64, error) {
	v := uint8(input[startOffset])
	if v == 0 {
		val.SetBool(false)
	} else if v == 1 {
		val.SetBool(true)
	} else {
		return 0, fmt.Errorf("expected 0 or 1 but received %d", v)
	}
	return startOffset + 1, nil
}

func unmarshalUint8(input []byte, val reflect.Value, startOffset uint64) (uint64, error) {
	val.SetUint(uint64(input[startOffset]))
	return startOffset + 1, nil
}

func unmarshalUint16(input []byte, val reflect.Value, startOffset uint64) (uint64, error) {
	offset := startOffset + 2
	buf := make([]byte, 2)
	copy(buf, input[startOffset:offset])
	val.SetUint(uint64(binary.LittleEndian.Uint16(buf)))
	return offset, nil
}

func unmarshalUint32(input []byte, val reflect.Value, startOffset uint64) (uint64, error) {
	offset := startOffset + 4
	buf := make([]byte, 4)
	copy(buf, input[startOffset:offset])
	val.SetUint(uint64(binary.LittleEndian.Uint32(buf)))
	return offset, nil
}

func unmarshalUint64(input []byte, val reflect.Value, startOffset uint64) (uint64, error) {
	offset := startOffset + 8
	buf := make([]byte, 8)
	copy(buf, input[startOffset:offset])
	val.SetUint(binary.LittleEndian.Uint64(buf))
	return offset, nil
}

func makeByteSliceUnmarshaler() (unmarshaler, error) {
	unmarshaler := func(input []byte, val reflect.Value, startOffset uint64) (uint64, error) {
		offset := startOffset + uint64(len(input))
		val.SetBytes(input[startOffset:offset])
		return offset, nil
	}
	return unmarshaler, nil
}

func makeBasicSliceUnmarshaler(typ reflect.Type) (unmarshaler, error) {
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, err
	}
	unmarshaler := func(input []byte, val reflect.Value, startOffset uint64) (uint64, error) {
		newVal := reflect.MakeSlice(val.Type(), 1, 1)
		reflect.Copy(newVal, val)
		val.Set(newVal)

		index := startOffset
		index, err = elemSSZUtils.unmarshaler(input, val.Index(0), index)
		if err != nil {
			return 0, fmt.Errorf("failed to unmarshal element of slice: %v", err)
		}
		elementSize := index - startOffset
		endOffset := uint64(len(input)) / elementSize

		newVal = reflect.MakeSlice(val.Type(), int(endOffset), int(endOffset))
		reflect.Copy(newVal, val)
		val.Set(newVal)
		i := uint64(1)
		for i < endOffset {
			index, err = elemSSZUtils.unmarshaler(input, val.Index(int(i)), index)
			if err != nil {
				return 0, fmt.Errorf("failed to unmarshal element of slice: %v", err)
			}
			i++
		}
		return index, nil
	}
	return unmarshaler, nil
}

func makeCompositeSliceUnmarshaler(typ reflect.Type) (unmarshaler, error) {
	elemType := typ.Elem()
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(elemType)
	if err != nil {
		return nil, err
	}
	unmarshaler := func(input []byte, val reflect.Value, startOffset uint64) (uint64, error) {
		// TODO: Limitation, creating a list of type pointers creates a list of nil values.
		newVal := reflect.MakeSlice(typ, 1, 1)
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
				return 0, errors.New("offset out of bounds")
			}
			nextIndex = currentIndex + uint64(BytesPerLengthOffset)
			if nextIndex == firstOffset {
				nextOffset = endOffset
			} else {
				nextOffsetVal := input[nextIndex : nextIndex+uint64(BytesPerLengthOffset)]
				nextOffset = startOffset + uint64(binary.LittleEndian.Uint32(nextOffsetVal))
			}
			if currentOffset > nextOffset {
				return 0, errors.New("offsets must be increasing")
			}
			// We grow the slice's size to accommodate a new element being unmarshald.
			newVal := reflect.MakeSlice(typ, i+1, i+1)
			reflect.Copy(newVal, val)
			val.Set(newVal)
			if _, err := elemSSZUtils.unmarshaler(input[currentOffset:nextOffset], val.Index(i), 0); err != nil {
				return 0, fmt.Errorf("failed to unmarshal element of slice: %v", err)
			}
			i++
			currentIndex = nextIndex
			currentOffset = nextOffset
		}
		return currentIndex, nil
	}
	return unmarshaler, nil
}

func makeBasicArrayUnmarshaler(typ reflect.Type) (unmarshaler, error) {
	elemType := typ.Elem()
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(elemType)
	if err != nil {
		return nil, err
	}
	unmarshaler := func(input []byte, val reflect.Value, startOffset uint64) (uint64, error) {
		i := 0
		index := startOffset
		size := val.Len()
		for i < size {
			index, err = elemSSZUtils.unmarshaler(input, val.Index(i), index)
			if err != nil {
				return 0, fmt.Errorf("failed to unmarshal element of array: %v", err)
			}
			i++
		}
		return index, nil
	}
	return unmarshaler, nil
}

func makeCompositeArrayUnmarshaler(typ reflect.Type) (unmarshaler, error) {
	elemType := typ.Elem()
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(elemType)
	if err != nil {
		return nil, err
	}
	unmarshaler := func(input []byte, val reflect.Value, startOffset uint64) (uint64, error) {
		currentIndex := startOffset
		nextIndex := currentIndex
		offsetVal := input[startOffset : startOffset+uint64(BytesPerLengthOffset)]
		firstOffset := startOffset + uint64(binary.LittleEndian.Uint32(offsetVal))
		currentOffset := firstOffset
		nextOffset := currentOffset
		endOffset := uint64(len(input))

		i := 0
		for currentIndex < firstOffset {
			if currentOffset > endOffset {
				return 0, errors.New("offset out of bounds")
			}
			nextIndex = currentIndex + uint64(BytesPerLengthOffset)
			if nextIndex == firstOffset {
				nextOffset = endOffset
			} else {
				nextOffsetVal := input[nextIndex : nextIndex+uint64(BytesPerLengthOffset)]
				nextOffset = startOffset + uint64(binary.LittleEndian.Uint32(nextOffsetVal))
			}
			if currentOffset > nextOffset {
				return 0, errors.New("offsets must be increasing")
			}
			if _, err := elemSSZUtils.unmarshaler(input[currentOffset:nextOffset], val.Index(i), 0); err != nil {
				return 0, fmt.Errorf("failed to unmarshal element of slice: %v", err)
			}
			i++
			currentIndex = nextIndex
			currentOffset = nextOffset
		}
		return currentIndex, nil
	}
	return unmarshaler, nil
}

func makeStructUnmarshaler(typ reflect.Type) (unmarshaler, error) {
	fields, err := structFields(typ)
	if err != nil {
		return nil, err
	}
	unmarshaler := func(input []byte, val reflect.Value, startOffset uint64) (uint64, error) {
		endOffset := uint64(len(input))
		currentIndex := startOffset
		nextIndex := currentIndex
		fixedSizes := make([]uint64, len(fields))

		for i := 0; i < len(fixedSizes); i++ {
			fixedSz := determineFixedSize(val.Field(i), val.Field(i).Type())
			if !isVariableSizeType(val.Field(i), val.Field(i).Type()) && (fixedSz > 0) {
				fixedSizes[i] = fixedSz
			} else {
				fixedSizes[i] = 0
			}
		}

		offsets := make([]uint64, 0)
		fixedEnd := uint64(0)
		for i, item := range fixedSizes {
			if item > 0 {
				fixedEnd += uint64(i) + item
			} else {
				offsetVal := input[i : i+BytesPerLengthOffset]
				offsets = append(offsets, startOffset+binary.LittleEndian.Uint64(offsetVal))
				fixedEnd += uint64(i + BytesPerLengthOffset)
			}
		}
		offsets = append(offsets, endOffset)

		offsetIndex := uint64(0)
		for i := 0; i < len(fields); i++ {
			f := fields[i]
			fieldSize := fixedSizes[i]
			if fieldSize > 0 {
				nextIndex = currentIndex + fieldSize
				if _, err := f.sszUtils.unmarshaler(input[currentIndex:nextIndex], val.Field(i), 0); err != nil {
					return 0, err
				}
				currentIndex = nextIndex

			} else {
				firstOff := offsets[offsetIndex]
				nextOff := offsets[offsetIndex+1]
				if _, err := f.sszUtils.unmarshaler(input[firstOff:nextOff], val.Field(i), 0); err != nil {
					return 0, err
				}
				offsetIndex++
				currentIndex += uint64(BytesPerLengthOffset)
			}
		}
		return 0, nil
	}
	return unmarshaler, nil
}

func makePtrUnmarshaler(typ reflect.Type) (unmarshaler, error) {
	elemType := typ.Elem()
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(elemType)
	if err != nil {
		return nil, err
	}
	unmarshaler := func(input []byte, val reflect.Value, startOffset uint64) (uint64, error) {
		elemSize, err := elemSSZUtils.unmarshaler(input, val.Elem(), startOffset)
		if err != nil {
			return 0, fmt.Errorf("failed to unmarshal to object pointed by pointer: %v", err)
		}
		return elemSize, nil
	}
	return unmarshaler, nil
}
