package ssz

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
)

// Unmarshal SSZ encoded data and output it into the object pointed by pointer val.
// Given a struct with the following fields, and some encoded bytes of type []byte,
// one can then unmarshal the bytes into a pointer of the struct as follows:
//  type exampleStruct1 struct {
//      Field1 uint8
//      Field2 []byte
//  }
//
//  var targetStruct exampleStruct1
//  if err := Unmarshal(encodedBytes, &targetStruct); err != nil {
//      return fmt.Errorf("failed to unmarshal: %v", err)
//  }
func Unmarshal(input []byte, val interface{}) error {
	if val == nil {
		return errors.New("cannot unmarshal into untyped, nil value")
	}
	rval := reflect.ValueOf(val)
	rtyp := rval.Type()
	// val must be a pointer, otherwise we refuse to unmarshal
	if rtyp.Kind() != reflect.Ptr {
		return errors.New("can only unmarshal into a pointer target")
	}
	if rval.IsNil() {
		return errors.New("cannot output to pointer of nil value")
	}
	sszUtils, err := cachedSSZUtils(rval.Elem().Type())
	if err != nil {
		return fmt.Errorf("could not initialize unmarshaler for type: %v, %v", rval.Elem().Type(), err)
	}
	if _, err = sszUtils.unmarshaler(input, rval.Elem(), 0); err != nil {
		return fmt.Errorf("could not unmarshal input into type: %v, %v", rval.Elem().Type(), err)
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
	case kind == reflect.Array && typ.Elem().Kind() == reflect.Uint8:
		return makeBasicArrayUnmarshaler(typ)
	case kind == reflect.Slice && isBasicTypeArray(typ.Elem(), typ.Elem().Kind()):
		return makeBasicSliceUnmarshaler(typ)
	case kind == reflect.Slice && isBasicType(typ.Elem().Kind()):
		return makeBasicSliceUnmarshaler(typ)
	case kind == reflect.Slice && !isVariableSizeType(typ.Elem()):
		return makeBasicSliceUnmarshaler(typ)
	case kind == reflect.Array && !isVariableSizeType(typ.Elem()):
		return makeBasicArrayUnmarshaler(typ)
	case kind == reflect.Slice:
		return makeCompositeSliceUnmarshaler(typ)
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
		if len(input) == 0 {
			newVal := reflect.MakeSlice(val.Type(), 0, 0)
			val.Set(newVal)
			return 0, nil
		}
		// If there are struct tags that specify a different type, we handle accordingly.
		if val.Type() != typ {
			sizes := []uint64{1}
			innerElement := typ.Elem()
			for {
				if innerElement.Kind() == reflect.Slice {
					sizes = append(sizes, 0)
					innerElement = innerElement.Elem()
				} else if innerElement.Kind() == reflect.Array {
					sizes = append(sizes, uint64(innerElement.Len()))
					innerElement = innerElement.Elem()
				} else {
					break
				}
			}
			// If the item is a slice, we grow it accordingly based on the size tags.
			result := growSliceFromSizeTags(val, sizes)
			reflect.Copy(result, val)
			val.Set(result)
		} else {
			growConcreteSliceType(val, val.Type(), 1)
		}

		index := startOffset
		index, err = elemSSZUtils.unmarshaler(input, val.Index(0), index)
		if err != nil {
			return 0, fmt.Errorf("failed to unmarshal element of slice: %v", err)
		}

		elementSize := index - startOffset
		endOffset := uint64(len(input)) / elementSize
		if val.Type() != typ {
			sizes := []uint64{endOffset}
			innerElement := typ.Elem()
			for {
				if innerElement.Kind() == reflect.Slice {
					sizes = append(sizes, 0)
					innerElement = innerElement.Elem()
				} else if innerElement.Kind() == reflect.Array {
					sizes = append(sizes, uint64(innerElement.Len()))
					innerElement = innerElement.Elem()
				} else {
					break
				}
			}
			// If the item is a slice, we grow it accordingly based on the size tags.
			result := growSliceFromSizeTags(val, sizes)
			reflect.Copy(result, val)
			val.Set(result)
		}
		i := uint64(1)
		for i < endOffset {
			if val.Type() == typ {
				growConcreteSliceType(val, val.Type(), int(i)+1)
			}
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
		if len(input) == 0 {
			newVal := reflect.MakeSlice(val.Type(), 0, 0)
			val.Set(newVal)
			return 0, nil
		}
		growConcreteSliceType(val, typ, 1)
		endOffset := uint64(len(input))

		currentIndex := startOffset
		nextIndex := currentIndex
		offsetVal := input[startOffset : startOffset+BytesPerLengthOffset]
		firstOffset := startOffset + uint64(binary.LittleEndian.Uint32(offsetVal))
		currentOffset := firstOffset
		nextOffset := currentOffset
		i := 0
		for currentIndex < firstOffset {
			nextIndex = currentIndex + BytesPerLengthOffset
			if nextIndex == firstOffset {
				nextOffset = endOffset
			} else {
				nextOffsetVal := input[nextIndex : nextIndex+BytesPerLengthOffset]
				nextOffset = startOffset + uint64(binary.LittleEndian.Uint32(nextOffsetVal))
			}
			// We grow the slice's size to accommodate a new element being unmarshaled.
			growConcreteSliceType(val, typ, i+1)
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
			if val.Index(i).Kind() == reflect.Ptr {
				instantiateConcreteTypeForElement(val.Index(i), typ.Elem().Elem())
			}
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
		offsetVal := input[startOffset : startOffset+BytesPerLengthOffset]
		firstOffset := startOffset + uint64(binary.LittleEndian.Uint32(offsetVal))
		currentOffset := firstOffset
		nextOffset := currentOffset
		endOffset := uint64(len(input))

		i := 0
		for currentIndex < firstOffset {
			nextIndex = currentIndex + BytesPerLengthOffset
			if nextIndex == firstOffset {
				nextOffset = endOffset
			} else {
				nextOffsetVal := input[nextIndex : nextIndex+BytesPerLengthOffset]
				nextOffset = startOffset + uint64(binary.LittleEndian.Uint32(nextOffsetVal))
			}
			if val.Index(i).Kind() == reflect.Ptr {
				instantiateConcreteTypeForElement(val.Index(i), typ.Elem().Elem())
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
			if !isVariableSizeType(fields[i].typ) {
				if val.Field(i).Kind() == reflect.Ptr {
					instantiateConcreteTypeForElement(val.Field(i), fields[i].typ.Elem())
				}
				concreteVal := val.Field(i)
				sszSizeTags, hasTags, err := parseSSZFieldTags(typ.Field(i))
				if err != nil {
					return 0, err
				}
				if hasTags {
					concreteType := inferFieldTypeFromSizeTags(typ.Field(i), sszSizeTags)
					concreteVal = reflect.New(concreteType).Elem()
					// If the item is a slice, we grow it accordingly based on the size tags.
					if val.Field(i).Kind() == reflect.Slice {
						result := growSliceFromSizeTags(val.Field(i), sszSizeTags)
						val.Field(i).Set(result)
					}
				}
				fixedSz := determineFixedSize(concreteVal, fields[i].typ)
				if fixedSz > 0 {
					fixedSizes[i] = fixedSz
				}
			} else {
				fixedSizes[i] = 0
			}
		}

		offsets := make([]uint64, 0)
		offsetIndexCounter := startOffset
		for _, item := range fixedSizes {
			if item > 0 {
				offsetIndexCounter += item
			} else {
				offsetVal := input[offsetIndexCounter : offsetIndexCounter+BytesPerLengthOffset]
				offsets = append(offsets, startOffset+uint64(binary.LittleEndian.Uint32(offsetVal)))
				offsetIndexCounter += BytesPerLengthOffset
			}
		}
		offsets = append(offsets, endOffset)
		offsetIndex := uint64(0)
		for i := 0; i < len(fields); i++ {
			f := fields[i]
			fieldSize := fixedSizes[i]
			if val.Field(i).Kind() == reflect.Ptr {
				instantiateConcreteTypeForElement(val.Field(i), fields[i].typ.Elem())
			}
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
				currentIndex += BytesPerLengthOffset
			}
		}
		return currentIndex, nil
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
