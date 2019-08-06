package types

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// UnboundedSSZFieldSizeMarker is the character used to specify a ssz field should have
// unbounded size, which is useful when describing slices of arrays such as [][32]byte.
// The ssz struct tag for such field type would be `ssz:"size=?,32"`. A question mark
// is chosen as the default value given its simplicity to represent unbounded size.
var UnboundedSSZFieldSizeMarker = "?"

type StructSSZ struct {
	hashCache map[string]interface{}
}

func NewStructSSZ() *StructSSZ {
	return &StructSSZ{
		hashCache: make(map[string]interface{}),
	}
}

func (b *StructSSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	if typ.Kind() == reflect.Ptr {
		if val.IsNil() {
			return startOffset, nil
		}
		return b.Marshal(val.Elem(), typ.Elem(), buf, startOffset)
	}
	fixedIndex := startOffset
	fixedLength := uint64(0)
	// For every field, we add up the total length of the items depending if they
	// are variable or fixed-size fields.
	for i := 0; i < typ.NumField(); i++ {
		// We skip protobuf related metadata fields.
		if strings.Contains(typ.Field(i).Name, "XXX") {
			continue
		}
		fType, err := determineFieldType(typ.Field(i))
		if err != nil {
			return 0, err
		}
		if isVariableSizeType(fType) {
			fixedLength += BytesPerLengthOffset
		} else {
			fixedLength += determineFixedSize(val.Field(i), fType)
		}
	}
	currentOffsetIndex := startOffset + fixedLength
	nextOffsetIndex := currentOffsetIndex
	for i := 0; i < typ.NumField(); i++ {
		// We skip protobuf related metadata fields.
		if strings.Contains(typ.Field(i).Name, "XXX") {
			continue
		}
		fType, err := determineFieldType(typ.Field(i))
		if err != nil {
			return 0, err
		}
		factory, err := SSZFactory(val.Field(i), fType)
		if err != nil {
			return 0, err
		}
		if !isVariableSizeType(fType) {
			fixedIndex, err = factory.Marshal(val.Field(i), fType, buf, fixedIndex)
			if err != nil {
				return 0, err
			}
		} else {
			nextOffsetIndex, err = factory.Marshal(val.Field(i), fType, buf, currentOffsetIndex)
			if err != nil {
				return 0, err
			}
			// Write the offset.
			offsetBuf := make([]byte, BytesPerLengthOffset)
			binary.LittleEndian.PutUint32(offsetBuf, uint32(currentOffsetIndex-startOffset))
			copy(buf[fixedIndex:fixedIndex+BytesPerLengthOffset], offsetBuf)

			// We increase the offset indices accordingly.
			currentOffsetIndex = nextOffsetIndex
			fixedIndex += BytesPerLengthOffset
		}
	}
	return currentOffsetIndex, nil
}

func (b *StructSSZ) Unmarshal(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	if typ.Kind() == reflect.Ptr {
		if val.IsNil() {
			return startOffset, nil
		}
		return b.Unmarshal(val.Elem(), typ.Elem(), input, startOffset)
	}
	endOffset := uint64(len(input))
	currentIndex := startOffset
	nextIndex := currentIndex
	fixedSizes := make([]uint64, typ.NumField())

	for i := 0; i < len(fixedSizes); i++ {
		fType, err := determineFieldType(typ.Field(i))
		if err != nil {
			return 0, err
		}
		if !isVariableSizeType(fType) {
			if val.Field(i).Kind() == reflect.Ptr {
				instantiateConcreteTypeForElement(val.Field(i), fType.Elem())
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
			fixedSz := determineFixedSize(concreteVal, fType)
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
	for i := 0; i < typ.NumField(); i++ {
		fieldSize := fixedSizes[i]
		fType, err := determineFieldType(typ.Field(i))
		if err != nil {
			return 0, err
		}
		if val.Field(i).Kind() == reflect.Ptr {
			instantiateConcreteTypeForElement(val.Field(i), fType.Elem())
		}
		factory, err := SSZFactory(val.Field(i), fType)
		if err != nil {
			return 0, err
		}
		if fieldSize > 0 {
			nextIndex = currentIndex + fieldSize
			if _, err := factory.Unmarshal(val.Field(i), fType, input[currentIndex:nextIndex], 0); err != nil {
				return 0, err
			}
			currentIndex = nextIndex

		} else {
			firstOff := offsets[offsetIndex]
			nextOff := offsets[offsetIndex+1]
			if _, err := factory.Unmarshal(val.Field(i), fType, input[firstOff:nextOff], 0); err != nil {
				return 0, err
			}
			offsetIndex++
			currentIndex += BytesPerLengthOffset
		}
	}
	return currentIndex, nil
}

func (b *StructSSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	if typ.Kind() == reflect.Ptr {
		if val.IsNil() {
			return [32]byte{}, nil
		}
		return b.Root(val.Elem(), typ.Elem(), maxCapacity)
	}
	numFields := typ.NumField()
	return b.FieldsHasher(val, typ, numFields)
}

func (b *StructSSZ) FieldsHasher(val reflect.Value, typ reflect.Type, numFields int) ([32]byte, error) {
	roots := make([][]byte, numFields)
	hashKey := make([]byte, 32*numFields)
	offset := 0
	var err error
	for i := 0; i < numFields; i++ {
		// We skip protobuf related metadata fields.
		if strings.Contains(typ.Field(i).Name, "XXX") {
			continue
		}
		fCapacity := determineFieldCapacity(typ.Field(i))
		if typ.Field(i).Name == "AggregationBits" || typ.Field(i).Name == "CustodyBits" {
			r, err := bitlistHasher(val.Field(i), fCapacity)
			if err != nil {
				return [32]byte{}, nil
			}
			roots[i] = r[:]
			copy(hashKey[offset:offset+32], r[:])
			offset += 32
			continue
		}
		fType, err := determineFieldType(typ.Field(i))
		if err != nil {
			return [32]byte{}, err
		}
		factory, err := SSZFactory(val.Field(i), fType)
		if err != nil {
			return [32]byte{}, err
		}
		r, err := factory.Root(val.Field(i), fType, fCapacity)
		if err != nil {
			return [32]byte{}, err
		}
		roots[i] = r[:]
		copy(hashKey[offset:offset+32], r[:])
		offset += 32
	}
	res := b.hashCache[string(hashKey)]
	if res != nil {
		return res.([32]byte), nil
	}
	root, err := bitwiseMerkleize(roots, uint64(numFields), uint64(numFields))
	if err != nil {
		return [32]byte{}, err
	}
	b.hashCache[string(hashKey)] = root
	return root, nil
}

func determineFieldType(field reflect.StructField) (reflect.Type, error) {
	fieldSizeTags, exists, err := parseSSZFieldTags(field)
	if err != nil {
		return nil, fmt.Errorf("could not parse ssz struct field tags: %v", err)
	}
	if exists {
		// If the field does indeed specify ssz struct tags, we infer the field's type.
		return inferFieldTypeFromSizeTags(field, fieldSizeTags), nil
	}
	return field.Type, nil
}

func determineFieldCapacity(field reflect.StructField) uint64 {
	tag, exists := field.Tag.Lookup("ssz-max")
	if !exists {
		return 0
	}
	val, err := strconv.ParseUint(tag, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

func parseSSZFieldTags(field reflect.StructField) ([]uint64, bool, error) {
	tag, exists := field.Tag.Lookup("ssz-size")
	if !exists {
		return nil, false, nil
	}
	items := strings.Split(tag, ",")
	sizes := make([]uint64, len(items))
	var err error
	for i := 0; i < len(items); i++ {
		// If a field is unbounded, we mark it with a size of 0.
		if items[i] == UnboundedSSZFieldSizeMarker {
			sizes[i] = 0
			continue
		}
		sizes[i], err = strconv.ParseUint(items[i], 10, 64)
		if err != nil {
			return nil, false, err
		}
	}
	return sizes, true, nil
}

func inferFieldTypeFromSizeTags(field reflect.StructField, sizes []uint64) reflect.Type {
	innerElement := field.Type.Elem()
	for i := 1; i < len(sizes); i++ {
		innerElement = innerElement.Elem()
	}
	currentType := innerElement
	for i := len(sizes) - 1; i >= 0; i-- {
		if sizes[i] == 0 {
			currentType = reflect.SliceOf(currentType)
		} else {
			currentType = reflect.ArrayOf(int(sizes[i]), currentType)
		}
	}
	return currentType
}
