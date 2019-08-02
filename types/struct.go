package types

import (
	"reflect"
)

type StructSSZ struct {
	hashCache map[string]interface{}
}

func NewStructSSZ() *StructSSZ {
	return &StructSSZ{
		hashCache: make(map[string]interface{}),
	}
}

func (b *StructSSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	roots := [][]byte{}
	hashKey := []byte{}
	var err error
	for i := 0; i < typ.NumField(); i++ {
		factory, _ := SSZFactory(val.Field(i), val.Field(i).Type())
		r, err := factory.Root(val.Field(i), val.Field(i).Type())
		if err != nil {
			return [32]byte{}, err
		}
		hashKey = append(hashKey, r[:]...)
		roots = append(roots, r[:])
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
