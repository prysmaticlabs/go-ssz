package types

import (
	"reflect"
)

type stringSSZ struct {
	hashCache map[string]interface{}
}

func newStringSSZ() *stringSSZ {
	return &stringSSZ{
		hashCache: make(map[string]interface{}),
	}
}

func (b *stringSSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	return [32]byte{}, nil
}

func (b *stringSSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	return 0, nil
}

func (b *stringSSZ) Unmarshal(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	return 0, nil
}
