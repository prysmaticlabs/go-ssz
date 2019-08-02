package ssz

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/minio/highwayhash"
)

type junkObject struct {
	D2Int64Slice [][]uint64
	Uint         uint64
	Int64Slice   []uint64
}

type tree struct {
	First  []*junkObject
	Second []*junkObject
}

func generateJunkObject(size uint64) []*junkObject {
	object := make([]*junkObject, size)
	for i := uint64(0); i < uint64(len(object)); i++ {
		d2Int64Slice := make([][]uint64, size)
		is := make([]uint64, size)
		uInt := uint64(time.Now().UnixNano())
		is[i] = i
		d2Int64Slice[i] = make([]uint64, size)
		for j := uint64(0); j < uint64(len(object)); j++ {
			d2Int64Slice[i][j] = i + j
		}
		object[i] = &junkObject{
			D2Int64Slice: d2Int64Slice,
			Uint:         uInt,
			Int64Slice:   is,
		}

	}
	return object
}

func TestCache_byHash(t *testing.T) {
	byteSl := [][]byte{{0, 0}, {1, 1}}
	useCache = false
	mr, err := HashTreeRoot(byteSl)
	if err != nil {
		t.Fatal(err)
	}
	marshaler, err := makeCompositeSliceMarshaler(reflect.TypeOf(byteSl))
	if err != nil {
		t.Fatal(err)
	}
	k, err := generateCacheKey(reflect.ValueOf(byteSl), marshaler, 0)
	if err != nil {
		t.Fatal(err)
	}
	h, _ := highwayhash.New(make([]byte, 32))
	if _, err := h.Write(k); err != nil {
		t.Fatal(err)
	}
	// We take the hash of the generate cache key.
	hs := h.Sum(nil)
	exists, _, err := hashCache.RootByEncodedHash(hs)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("Expected block info not to exist in empty cache")
	}
	useCache = true
	if _, err := HashTreeRoot(byteSl); err != nil {
		t.Fatal(err)
	}
	exists, fetchedInfo, err := hashCache.RootByEncodedHash(hs)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("Expected blockInfo to exist")
	}
	if !bytes.Equal(mr[:], fetchedInfo.MerkleRoot) {
		t.Errorf(
			"Expected fetched info number to be %v, got %v",
			mr,
			fetchedInfo.MerkleRoot,
		)
	}
	if !bytes.Equal(fetchedInfo.Hash, hs) {
		t.Errorf(
			"Expected fetched info hash to be %v, got %v",
			hs,
			fetchedInfo.Hash,
		)
	}
}

func BenchmarkHashWithoutCache(b *testing.B) {
	useCache = false
	First := generateJunkObject(100)
	HashTreeRoot(&tree{First: First, Second: First})
	for n := 0; n < b.N; n++ {
		HashTreeRoot(&tree{First: First, Second: First})
	}
}

func BenchmarkHashWithCache(b *testing.B) {
	useCache = true
	First := generateJunkObject(100)
	type tree struct {
		First  []*junkObject
		Second []*junkObject
	}
	HashTreeRoot(&tree{First: First, Second: First})
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		HashTreeRoot(&tree{First: First, Second: First})
	}
}
