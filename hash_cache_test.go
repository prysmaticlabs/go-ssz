package ssz

import (
	"bytes"
	"log"
	"reflect"
	"testing"
	"time"
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
	hs, err := HashedEncoding(reflect.ValueOf(byteSl))
	if err != nil {
		t.Fatal(err)
	}
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
	if fetchedInfo.Hash != hs {
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

func TestBlockCache_maxSize(t *testing.T) {
	maxCacheSize := int64(10000)
	cache := newHashCache(maxCacheSize)
	for i := uint64(0); i < uint64(maxCacheSize+1025); i++ {

		if err := cache.AddRoot(toBytes32(bytes4(i)), []byte{1}); err != nil {
			t.Fatal(err)
		}
	}
	log.Printf(
		"hash cache key size is %d, itemcount is %d",
		maxCacheSize,
		cache.hashCache.ItemCount(),
	)
	if int64(cache.hashCache.ItemCount()) > maxCacheSize {
		t.Errorf(
			"Expected hash cache key size to be %d, got %d",
			maxCacheSize,
			cache.hashCache.ItemCount(),
		)
	}
}
