package types

import (
	"reflect"
	"sync"

	"github.com/karlseguin/ccache"
)

// BasicArraySizeCache for HashTreeRoot.
const RootsArraySizeCache = 100000

type rootsArraySSZ struct {
	hashCache *ccache.Cache
	lock      sync.Mutex
}

func newRootsArraySSZ() *rootsArraySSZ {
	return &rootsArraySSZ{
		hashCache: ccache.New(ccache.Configure().MaxSize(BasicArraySizeCache)),
	}
}

func (a *rootsArraySSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	return [32]byte{}, nil
}

func (a *rootsArraySSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	index := startOffset
	return index, nil
}

func (a *rootsArraySSZ) Unmarshal(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	return 0, nil
}
