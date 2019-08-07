package types

import (
	"bytes"
	"reflect"
	"sync"
)

type basicArraySSZ struct {
	hashCache map[string]interface{}
	lock      sync.Mutex
}

func newBasicArraySSZ() *basicArraySSZ {
	return &basicArraySSZ{
		hashCache: make(map[string]interface{}),
	}
}

func (b *basicArraySSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	numItems := val.Len()
	hashKey := make([]byte, BytesPerChunk*numItems)
	leaves := make([][]byte, numItems)
	elemKind := typ.Elem().Kind()
	offset := 0
	var factory SSZAble
	var err error
	if numItems > 0 {
		factory, err = SSZFactory(val.Index(0), typ.Elem())
		if err != nil {
			return [32]byte{}, err
		}
	}
	for i := 0; i < numItems; i++ {
		// If we are marshaling an byte array of length 32, we shortcut the computations and
		// simply return it as an identity root.
		if elemKind == reflect.Array && typ.Elem().Elem().Kind() == reflect.Uint8 && val.Index(i).Len() == 32 {
			leaves[i] = val.Index(i).Bytes()
			copy(hashKey[offset:offset+32], leaves[i])
			offset += 32
			continue
		}
		r, err := factory.Root(val.Index(i), typ.Elem(), 0)
		if err != nil {
			return [32]byte{}, err
		}
		leaves[i] = r[:]
		copy(hashKey[offset:offset+32], r[:])
		offset += 32
	}
	if !bytes.Equal(hashKey, make([]byte, BytesPerChunk*numItems)) {
		b.lock.Lock()
		res := b.hashCache[string(hashKey)]
		b.lock.Unlock()
		if res != nil {
			return res.([32]byte), nil
		}
	}
	chunks, err := pack(leaves)
	if err != nil {
		return [32]byte{}, err
	}
	root, err := bitwiseMerkleize(chunks, uint64(len(chunks)), uint64(len(chunks)))
	if err != nil {
		return [32]byte{}, err
	}
	if !bytes.Equal(hashKey, make([]byte, BytesPerChunk*numItems)) {
		b.lock.Lock()
		b.hashCache[string(hashKey)] = root
		b.lock.Unlock()
	}
	return root, nil
}

func (b *basicArraySSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	index := startOffset
	var err error
	if val.Len() == 0 {
		return index, nil
	}
	factory, err := SSZFactory(val.Index(0), typ.Elem())
	if err != nil {
		return 0, err
	}
	for i := 0; i < val.Len(); i++ {
		index, err = factory.Marshal(val.Index(i), typ.Elem(), buf, index)
		if err != nil {
			return 0, err
		}
	}
	return index, nil
}

func (b *basicArraySSZ) Unmarshal(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	i := 0
	index := startOffset
	size := val.Len()
	var err error
	var factory SSZAble
	for i < size {
		if val.Index(i).Kind() == reflect.Ptr {
			instantiateConcreteTypeForElement(val.Index(i), typ.Elem().Elem())
			factory, err = SSZFactory(val.Index(i), typ.Elem().Elem())
			if err != nil {
				return 0, err
			}
		} else {
			factory, err = SSZFactory(val.Index(i), typ.Elem())
			if err != nil {
				return 0, err
			}
		}
		index, err = factory.Unmarshal(val.Index(i), typ.Elem(), input, index)
		if err != nil {
			return 0, err
		}
		i++
	}
	return index, nil
}
