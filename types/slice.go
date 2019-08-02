package types

import (
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
