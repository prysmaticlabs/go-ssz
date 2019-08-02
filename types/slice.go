package types

import (
	"reflect"

	"github.com/karlseguin/ccache"
)

type BasicSliceSSZ struct {
	hashCache *ccache.Cache
}

type CompositeSliceSSZ struct {
	hashCache *ccache.Cache
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

// TODO: Add max capacity
func (b *BasicSliceSSZ) Root(val reflect.Value, typ reflect.Type) ([32]byte, error) {
	elemSize := uint64(0)
	if isBasicType(typ.Elem().Kind()) {
		elemSize = determineFixedSize(val, typ.Elem())
	} else {
		elemSize = 32
	}
	limit := (maxCapacity*elemSize + 31) / 32
	if limit == 0 {
		limit = 1
	}

	var leaves [][]byte
	var err error
	for i := 0; i < val.Len(); i++ {
		if isBasicType(val.Index(i).Kind()) {
			innerBuf := make([]byte, elemSize)
			if _, err = recursiveMarshal(val.Index(i), typ.Elem(), innerBuf, 0); err != nil {
				return [32]byte{}, err
			}
			leaves = append(leaves, innerBuf)
		} else {
			r, err := recursiveHasher(val.Index(i), typ.Elem(), 0)
			if err != nil {
				return [32]byte{}, err
			}
			leaves = append(leaves, r[:])
		}
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

