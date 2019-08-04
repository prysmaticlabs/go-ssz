package types

import (
	"bytes"
	"encoding/binary"
	"reflect"
)

type BasicSliceSSZ struct {
	hashCache map[string]interface{}
}

type CompositeSliceSSZ struct {
	hashCache map[string]interface{}
}

type BasicArraySSZ struct {
	hashCache map[string]interface{}
}

type CompositeArraySSZ struct {
	hashCache map[string]interface{}
}

type BitlistSSZ struct {
	hashCache map[string]interface{}
}

func NewBasicSliceSSZ() *BasicSliceSSZ {
	return &BasicSliceSSZ{
		hashCache: make(map[string]interface{}),
	}
}

func NewBasicArraySSZ() *BasicArraySSZ {
	return &BasicArraySSZ{
		hashCache: make(map[string]interface{}),
	}
}

func NewCompositeArraySSZ() *CompositeArraySSZ {
	return &CompositeArraySSZ{
		hashCache: make(map[string]interface{}),
	}
}

func NewCompositeSliceSSZ() *CompositeSliceSSZ {
	return &CompositeSliceSSZ{
		hashCache: make(map[string]interface{}),
	}
}

func (b *CompositeSliceSSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	output := make([]byte, 32)
	if val.Len() == 0 && maxCapacity == 0 {
		merkleRoot, err := bitwiseMerkleize([][]byte{}, 0, true /* has limit */)
		if err != nil {
			return [32]byte{}, err
		}
		return mixInLength(merkleRoot, output), nil
	}
	factory, _ := SSZFactory(val.Index(0), typ.Elem())
	numItems := val.Len()
	roots := make([][]byte, numItems)
	for i := 0; i < numItems; i++ {
		r, err := factory.Root(val.Index(i), typ.Elem(), 0)
		if err != nil {
			return [32]byte{}, err
		}
		roots[i] = r[:]
	}
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, uint64(val.Len())); err != nil {
		return [32]byte{}, err
	}
	copy(output, buf.Bytes())
	objLen := maxCapacity
	if maxCapacity == 0 {
		objLen = uint64(val.Len())
	}
	merkleRoot, err := bitwiseMerkleize(roots, objLen, true /* has limit */)
	if err != nil {
		return [32]byte{}, err
	}
	return mixInLength(merkleRoot, output), nil
}

func (b *CompositeSliceSSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	return 0, nil
}

func (b *BasicSliceSSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	var factory SSZAble
	elemSize := determineFixedSize(val, typ.Elem())
	limit := (maxCapacity*elemSize + 31) / 32
	numItems := val.Len()
	if numItems > 0 {
		factory, _ = SSZFactory(val.Index(0), typ.Elem())
		// TODO: Abstract into a chunk count function.
	}
	if limit == 0 {
		limit = 1
	}
	leaves := make([][]byte, numItems)
	var err error
	for i := 0; i < numItems; i++ {
		innerBuf := make([]byte, elemSize)
		if _, err = factory.Marshal(val.Index(i), typ.Elem(), innerBuf, 0); err != nil {
			return [32]byte{}, err
		}
		leaves[i] = innerBuf
	}
	chunks, err := pack(leaves)
	if err != nil {
		return [32]byte{}, err
	}
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, uint64(val.Len())); err != nil {
		return [32]byte{}, err
	}
	output := make([]byte, 32)
	copy(output, buf.Bytes())
	merkleRoot, err := bitwiseMerkleize(chunks, limit, true /* has limit */)
	if err != nil {
		return [32]byte{}, err
	}
	return mixInLength(merkleRoot, output), nil
}

func (b *BasicSliceSSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	return 0, nil
}

func (b *CompositeArraySSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	factory, _ := SSZFactory(val.Index(0), typ.Elem())
	numItems := val.Len()
	roots := make([][]byte, numItems)
	hashKey := []byte{}
	limit := uint64(typ.Len())
	var err error
	for i := 0; i < val.Len(); i++ {
		r, err := factory.Root(val.Index(i), typ.Elem(), 0)
		if err != nil {
			return [32]byte{}, err
		}
		hashKey = append(hashKey, r[:]...)
		roots[i] = r[:]
	}

	res := b.hashCache[string(hashKey)]
	if res != nil {
		return res.([32]byte), nil
	}
	root, err := bitwiseMerkleize(roots, limit, true /* has limit */)
	if err != nil {
		return [32]byte{}, err
	}
	b.hashCache[string(hashKey)] = root
	return root, nil
}

func (b *CompositeArraySSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	return 0, nil
}

func (b *BasicArraySSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	numItems := val.Len()
	hashKey := make([]byte, 0)
	leaves := make([][]byte, numItems)
	var err error
	for i := 0; i < numItems; i++ {
		if typ.Elem().Kind() == reflect.Slice || typ.Elem().Kind() == reflect.Array {
			if val.Index(i).Len() == 32 {
				leaves[i] = val.Index(i).Interface().([]byte)
				continue
			}
		}
		r, err := basicSSZ.Root(val.Index(i), typ.Elem(), 0)
		if err != nil {
			return [32]byte{}, err
		}
		leaves[i] = r[:]
		hashKey = append(hashKey, r[:]...)
	}
	res := b.hashCache[string(hashKey)]
	if res != nil {
		return res.([32]byte), nil
	}
	root, err := bitwiseMerkleize(leaves, 1, false /* has limit */)
	if err != nil {
		return [32]byte{}, err
	}
	b.hashCache[string(hashKey)] = root
	return root, nil
}

func (b *BasicArraySSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	return 0, nil
}
