package types

import (
	"bytes"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/karlseguin/ccache"
	"github.com/minio/highwayhash"
	"github.com/protolambda/zssz/merkle"
)

// RootsArraySizeCache for hash tree root.
const RootsArraySizeCache = 100000

type rootsArraySSZ struct {
	hashCache    *ccache.Cache
	lock         sync.Mutex
	cachedLeaves map[string][][]byte
	layers       map[string][][][]byte
}

func newRootsArraySSZ() *rootsArraySSZ {
	return &rootsArraySSZ{
		hashCache:    ccache.New(ccache.Configure().MaxSize(RootsArraySizeCache)),
		cachedLeaves: make(map[string][][]byte),
		layers:       make(map[string][][][]byte),
	}
}

func (a *rootsArraySSZ) Root(val reflect.Value, typ reflect.Type, fieldName string, maxCapacity uint64) ([32]byte, error) {
	numItems := typ.Len()
	log.Println(fieldName)
	if fieldName != "" {
		if _, ok := a.layers[fieldName]; !ok {
			depth := merkle.GetDepth(uint64(numItems))
			a.layers[fieldName] = make([][][]byte, depth+1)
		}
	}
	hashKeyElements := make([]byte, BytesPerChunk*numItems)
	emptyKey := highwayhash.Sum(hashKeyElements, fastSumHashKey[:])
	offset := 0
	leaves := make([][]byte, numItems)
	changedIndices := make([]int, 0)
	for i := 0; i < numItems; i++ {
		// If we are marshaling an byte array of length 32, we shortcut the computations and
		// simply return it as an identity root.
		item := val.Index(i).Interface().([32]byte)
		leaves[i] = item[:]
		copy(hashKeyElements[offset:offset+32], leaves[i])
		offset += 32
		if fieldName != "" {
			if _, ok := a.cachedLeaves[fieldName]; ok {
				if !bytes.Equal(leaves[i], a.cachedLeaves[fieldName][i]) {
					changedIndices = append(changedIndices, i)
				}
			}
		}
	}
	chunks := leaves
	if len(changedIndices) > 0 {
		var rt [32]byte
		for i := 0; i < len(changedIndices); i++ {
			rt = a.recomputeRoot(changedIndices[i], chunks, fieldName)
		}
		return rt, nil
	}
	hashKey := highwayhash.Sum(hashKeyElements, fastSumHashKey[:])
	if hashKey != emptyKey {
		a.lock.Lock()
		res := a.hashCache.Get(string(hashKey[:]))
		a.lock.Unlock()
		if res != nil && res.Value() != nil {
			return res.Value().([32]byte), nil
		}
	}
	root := a.merkleize(chunks, fieldName)
	if fieldName != "" {
		a.cachedLeaves[fieldName] = leaves
	}
	if hashKey != emptyKey {
		a.lock.Lock()
		a.hashCache.Set(string(hashKey[:]), root, time.Hour)
		a.lock.Unlock()
	}
	return root, nil
}

func (a *rootsArraySSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	index := startOffset
	return index, nil
}

func (a *rootsArraySSZ) Unmarshal(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	return 0, nil
}

func (a *rootsArraySSZ) recomputeRoot(idx int, chunks [][]byte, fieldName string) [32]byte {
	root := chunks[idx]
	for i := 0; i < len(a.layers[fieldName])-1; i++ {
		subIndex := (uint64(idx) / (1 << uint64(i))) ^ 1
		isLeft := uint64(idx) / (1 << uint64(i))
		parentIdx := uint64(idx) / (1 << uint64(i+1))
		item := a.layers[fieldName][i][subIndex]
		if isLeft%2 != 0 {
			parentHash := hash(append(item, root...))
			root = parentHash[:]
		} else {
			parentHash := hash(append(root, item...))
			root = parentHash[:]
		}
		a.layers[fieldName][i+1][parentIdx] = root
	}
	return toBytes32(root)
}

func (a *rootsArraySSZ) merkleize(chunks [][]byte, fieldName string) [32]byte {
	if len(chunks) == 1 {
		var root [32]byte
		copy(root[:], chunks[0])
		return root
	}
	for !isPowerOf2(len(chunks)) {
		chunks = append(chunks, make([]byte, BytesPerChunk))
	}
	hashLayer := chunks
	if fieldName != "" {
		a.layers[fieldName][0] = hashLayer
	}
	// We keep track of the hash layers of a Merkle trie until we reach
	// the top layer of length 1, which contains the single root element.
	//        [Root]      -> Top layer has length 1.
	//    [E]       [F]   -> This layer has length 2.
	// [A]  [B]  [C]  [D] -> The bottom layer has length 4 (needs to be a power of two.
	i := 1
	for len(hashLayer) > 1 {
		layer := [][]byte{}
		for i := 0; i < len(hashLayer); i += 2 {
			hashedChunk := hash(append(hashLayer[i], hashLayer[i+1]...))
			layer = append(layer, hashedChunk[:])
		}
		hashLayer = layer
		if fieldName != "" {
			a.layers[fieldName][i] = hashLayer
		}
		i++
	}
	var root [32]byte
	copy(root[:], hashLayer[0])
	return root
}

func isPowerOf2(n int) bool {
	return n != 0 && (n&(n-1)) == 0
}
