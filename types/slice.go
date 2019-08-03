package types

import (
	"bytes"
	"encoding/binary"
	"reflect"

	"github.com/prysmaticlabs/go-bitfield"
)

type BasicSliceSSZ struct {
	hashCache map[string]interface{}
}

type CompositeSliceSSZ struct {
	hashCache map[string]interface{}
}

type CompositeArraySSZ struct {
	hashCache map[string]interface{}
}

func NewBasicSliceSSZ() *BasicSliceSSZ {
	return &BasicSliceSSZ{
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
	factory, _ := SSZFactory(val.Elem(), typ.Elem())
	limit := uint64(val.Len())
	if limit == 0 {
		limit = 1
	}
	var roots [][]byte
	for i := 0; i < val.Len(); i++ {
		r, err := factory.Root(val.Index(i), typ.Elem(), 0)
		if err != nil {
			return [32]byte{}, err
		}
		roots = append(roots, r[:])
	}
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, uint64(val.Len())); err != nil {
		return [32]byte{}, err
	}
	output := make([]byte, 32)
	copy(output, buf.Bytes())
	merkleRoot, err := bitwiseMerkleize(roots, limit, true /* has limit */)
	if err != nil {
		return [32]byte{}, err
	}
	return mixInLength(merkleRoot, output), nil
}

func (b *CompositeSliceSSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	return 0, nil
}

func (b *BasicSliceSSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	if _, ok := val.Interface().(bitfield.Bitlist); ok {
		return bitlistHasher(val, maxCapacity)
	}
	factory, _ := SSZFactory(val.Elem(), typ.Elem())
	// TODO: Abstract into a chunk count function.
	elemSize := determineFixedSize(val, typ.Elem())
	limit := (maxCapacity*elemSize + 31) / 32
	if limit == 0 {
		limit = 1
	}
	var leaves [][]byte
	var err error
	for i := 0; i < val.Len(); i++ {
		innerBuf := make([]byte, elemSize)
		if _, err = factory.Marshal(val.Index(i), typ.Elem(), innerBuf, 0); err != nil {
			return [32]byte{}, err
		}
		leaves = append(leaves, innerBuf)
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
	roots := [][]byte{}
	hashKey := []byte{}
	limit := uint64(typ.Len())
	var err error
	for i := 0; i < val.Len(); i++ {
		r, err := factory.Root(val.Index(i), typ.Elem(), 0)
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
