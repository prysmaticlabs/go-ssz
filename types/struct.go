package types

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

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
		fmt.Printf("%s has root %#x\n", typ.Field(i).Name, r)
		copy(hashKey[offset:offset+32], r[:])
		offset += 32
	}
	res := b.hashCache[string(hashKey)]
	if res != nil {
		return res.([32]byte), nil
	}
	root, err := NewMerkleize(roots, uint64(numFields), uint64(numFields), true)
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
		if items[i] == "?" {
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
