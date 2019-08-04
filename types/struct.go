package types

import (
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
	return 0, nil
}

func (b *StructSSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	numFields := typ.NumField()
	roots := make([][]byte, numFields)
	hashKey := []byte{}
	var err error
	//isState := strings.Contains(typ.Name(), "BeaconState")
	for i := 0; i < numFields; i++ {
		fCapacity := determineFieldCapacity(typ.Field(i))
		if typ.Field(i).Name == "AggregationBits" {
			r, err := bitlistHasher(val.Field(i), fCapacity)
			if err != nil {
				return [32]byte{}, nil
			}
			hashKey = append(hashKey, r[:]...)
			roots[i] = r[:]
			continue
		}
		fType, err := determineFieldType(typ.Field(i))
		if err != nil {
			return [32]byte{}, err
		}
		factory, _ := SSZFactory(val.Field(i), fType)
		//start := time.Now()
		r, err := factory.Root(val.Field(i), fType, fCapacity)
		if err != nil {
			return [32]byte{}, err
		}
		//end := time.Now()
		//if isState {
		//	fv := reflect.ValueOf(factory)
		//	fmt.Printf("%s took time %v with factory %v\n", typ.Field(i).Name, end.Sub(start), fv.Type())
		//}
		hashKey = append(hashKey, r[:]...)
		roots[i] = r[:]
	}
	res := b.hashCache[string(hashKey)]
	if res != nil {
		return res.([32]byte), nil
	}
	root, err := bitwiseMerkleize(roots, uint64(typ.NumField()), true /* has limit */)
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
