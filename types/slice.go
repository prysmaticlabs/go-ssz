package types

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"time"

	"github.com/karlseguin/ccache"
)

type CompositeArraySSZ struct {
	hashCache *ccache.Cache
}

type BasicSliceSSZ struct {
	hashCache *ccache.Cache
}

type CompositeSliceSSZ struct {
	hashCache *ccache.Cache
}

func NewCompositeArraySSZ() *CompositeArraySSZ {
	return &CompositeArraySSZ{
		hashCache: ccache.New(ccache.Configure().MaxSize(100000)),
	}
}

func NewBasicSliceSSZ() *BasicSliceSSZ {
	return &BasicSliceSSZ{
		hashCache: ccache.New(ccache.Configure().MaxSize(100000)),
	}
}

func NewCompositeSliceSSZ() *CompositeSliceSSZ {
	return &CompositeSliceSSZ{
		hashCache: ccache.New(ccache.Configure().MaxSize(100000)),
	}
}

func (b *BasicSliceSSZ) Root(val reflect.Value, typ reflect.Type) ([32]byte, error) {
	factory, _ := SSZFactory(val.Elem(), typ.Elem())
	elemSize := determineFixedSize(val, typ.Elem())
	maxCapacity := uint64(100)
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

func (b *CompositeArraySSZ) Root(val reflect.Value, typ reflect.Type) ([32]byte, error) {
	factory, _ := SSZFactory(val.Elem(), typ.Elem())
	roots := [][]byte{}
	hashKey := []byte{}
	limit := uint64(10000)
	var err error
	for i := 0; i < val.Len(); i++ {
		r, err := factory.Root(val.Index(i), typ.Elem())
		if err != nil {
			return [32]byte{}, err
		}
		hashKey = append(hashKey, r[:]...)
		roots = append(roots, r[:])
	}

	res := b.hashCache.Get(string(hashKey))
	if res != nil {
		return res.Value().([32]byte), nil
	}

	chunks, err := pack(roots)
	if err != nil {
		return [32]byte{}, err
	}
	if val.Len() == 0 {
		chunks = [][]byte{}
	}
	root, err := bitwiseMerkleize(chunks, limit, true /* has limit */)
	if err != nil {
		return [32]byte{}, err
	}
	b.hashCache.Set(string(hashKey), root, time.Hour)
	return root, nil
}
