package types

import (
	"reflect"
	"time"

	"github.com/karlseguin/ccache"
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

func (b *BasicSliceSSZ) Root(val reflect.Value, typ reflect.Type) ([32]byte, error) {
	return [32]byte{}, nil
	//factory, _ := SSZFactory(val.Elem(), typ.Elem())
	//elemSize := determineFixedSize(val, typ.Elem())
	//maxCapacity := uint64(100)
	//limit := (maxCapacity*elemSize + 31) / 32
	//if limit == 0 {
	//	limit = 1
	//}
	//var leaves [][]byte
	//var err error
	//for i := 0; i < val.Len(); i++ {
	//	innerBuf := make([]byte, elemSize)
	//	if _, err = factory.Marshal(val.Index(i), typ.Elem(), innerBuf, 0); err != nil {
	//		return [32]byte{}, err
	//	}
	//	leaves = append(leaves, innerBuf)
	//}
	//chunks, err := pack(leaves)
	//if err != nil {
	//	return [32]byte{}, err
	//}
	//buf := new(bytes.Buffer)
	//if err := binary.Write(buf, binary.LittleEndian, uint64(val.Len())); err != nil {
	//	return [32]byte{}, err
	//}
	//output := make([]byte, 32)
	//copy(output, buf.Bytes())
	//merkleRoot, err := bitwiseMerkleize(chunks, limit, true /* has limit */)
	//if err != nil {
	//	return [32]byte{}, err
	//}
	//return mixInLength(merkleRoot, output), nil
}

func (b *CompositeArraySSZ) Root(val reflect.Value, typ reflect.Type) ([32]byte, error) {
	factory, _ := SSZFactory(val.Index(0), typ.Elem())
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
	root, err := bitwiseMerkleize(roots, limit, true /* has limit */)
	if err != nil {
		return [32]byte{}, err
	}
	b.hashCache.Set(string(hashKey), root, time.Hour)
	return root, nil
}
