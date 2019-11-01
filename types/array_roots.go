package types

import (
	"bytes"
	"fmt"
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
	// We make sure to look into the cache only if a field name is provided, that is,
	// if this function is called when calling HashTreeRoot on a struct type that has
	// a field which is an array of roots. An example is:
	//
	// type BeaconState struct {
	//   BlockRoots [2048][32]byte
	// }
	//
	// which would allow us to look into the cache by the field "BlockRoots".
	if enableCache && fieldName != "" {
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
		var item [32]byte
		if res, ok := val.Index(i).Interface().([32]byte); ok {
			item = res
		} else if res, ok := val.Index(i).Interface().([]byte); ok {
			item = toBytes32(res)
		} else {
			return [32]byte{}, fmt.Errorf("expected array or slice of len 32, received %v", val.Index(i))
		}
		leaves[i] = item[:]
		copy(hashKeyElements[offset:offset+32], leaves[i])
		offset += 32
		if enableCache && fieldName != "" {
			if _, ok := a.cachedLeaves[fieldName]; ok {
				if !bytes.Equal(leaves[i], a.cachedLeaves[fieldName][i]) {
					changedIndices = append(changedIndices, i)
				}
			}
		}
	}
	chunks := leaves
	// Recompute the root from the modified branches from the previous call
	// to this function.
	if len(changedIndices) > 0 {
		var rt [32]byte
		for i := 0; i < len(changedIndices); i++ {
			rt = a.recomputeRoot(changedIndices[i], chunks, fieldName)
		}
		return rt, nil
	}
	hashKey := highwayhash.Sum(hashKeyElements, fastSumHashKey[:])
	if enableCache && hashKey != emptyKey {
		a.lock.Lock()
		res := a.hashCache.Get(string(hashKey[:]))
		a.lock.Unlock()
		if res != nil && res.Value() != nil {
			return res.Value().([32]byte), nil
		}
	}
	root := a.merkleize(chunks, fieldName)
	if enableCache && fieldName != "" {
		a.cachedLeaves[fieldName] = leaves
	}
	if enableCache && hashKey != emptyKey {
		a.lock.Lock()
		a.hashCache.Set(string(hashKey[:]), root, time.Hour)
		a.lock.Unlock()
	}
	return root, nil
}

func (a *rootsArraySSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	index := startOffset
	if val.Len() == 0 {
		return index, nil
	}
	for i := 0; i < val.Len(); i++ {
		var item [32]byte
		if res, ok := val.Index(i).Interface().([32]byte); ok {
			item = res
		} else if res, ok := val.Index(i).Interface().([]byte); ok {
			item = toBytes32(res)
		} else {
			return 0, fmt.Errorf("expected array or slice of len 32, received %v", val.Index(i))
		}
		copy(buf[index:index+uint64(len(item))], item[:])
		index += uint64(len(item))
	}
	return index, nil
}

func (a *rootsArraySSZ) Unmarshal(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	i := 0
	index := startOffset
	for i < val.Len() {
		val.Index(i).SetBytes(input[index : index+uint64(32)])
		index += uint64(32)
		i++
	}
	return index, nil
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
		// Update the cached layers at the parent index.
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
	if enableCache && fieldName != "" {
		a.layers[fieldName][0] = hashLayer
	}
	// We keep track of the hash layers of a Merkle trie until we reach
	// the top layer of length 1, which contains the single root element.
	//        [Root]      -> Top layer has length 1.
	//    [E]       [F]   -> This layer has length 2.
	// [A]  [B]  [C]  [D] -> The bottom layer has length 4 (needs to be a power of two).
	i := 1
	for len(hashLayer) > 1 {
		layer := [][]byte{}
		for i := 0; i < len(hashLayer); i += 2 {
			hashedChunk := hash(append(hashLayer[i], hashLayer[i+1]...))
			layer = append(layer, hashedChunk[:])
		}
		hashLayer = layer
		if enableCache && fieldName != "" {
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
