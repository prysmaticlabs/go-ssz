package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/karlseguin/ccache"
	"github.com/minio/highwayhash"
	"github.com/minio/sha256-simd"
	pb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
)

const BytesPerChunk = 32

type marshaler func(val reflect.Value, buf []byte, startOffset uint64) (uint64, error)

type BeaconState struct {
	BlockRoots [65536][32]byte
}

var (
	typesMap = make(map[reflect.Type]marshaler)
	fastSumHashKey = toBytes32([]byte("hash_fast_sum64_key"))
)

func init() {
	check := pb.Checkpoint{
		Epoch: 1,
		Root: make([]byte, 32),
	}
	copy(check.Root, "12345678910")
	val := reflect.ValueOf(check)
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		if strings.Contains(typ.Field(i).Name, "XXX_") {
			continue
		}
		if val.Field(i).Type().Kind() == reflect.Uint64 {
			typesMap[val.Field(i).Type()] = marshalUint64
		} else if val.Field(i).Type().Kind() == reflect.Slice {
			typesMap[val.Field(i).Type()] = marshalByteArray
		}
	}
}

func main() {
	bs := BeaconState{
		BlockRoots: [65536][32]byte{},
	}
	for i := 0; i < len(bs.BlockRoots); i++ {
		bs.BlockRoots[i] = [32]byte{1, 2, 3}
	}
	hasher := &structHasher{
		layers: make([][][]byte, 17),
		hashCache: ccache.New(ccache.Configure().MaxSize(10000)),
	}
	start := time.Now()
	res, err := hasher.fastHashTreeRoot(bs)
	if err != nil {
		panic(err)
	}
	end := time.Now()
	log.Printf("%#x in %v", res, end.Sub(start))
	start = time.Now()
	res, err = hasher.fastHashTreeRoot(bs)
	if err != nil {
		panic(err)
	}
	end = time.Now()
	log.Printf("%#x in %v", res, end.Sub(start))
	// Change 1 element.
	bs.BlockRoots[len(bs.BlockRoots)-1] = [32]byte{4, 5, 6}
	start = time.Now()
	res, err = hasher.fastHashTreeRoot(bs)
	if err != nil {
		panic(err)
	}
	end = time.Now()
	log.Printf("%#x in %v", res, end.Sub(start))
}

func fastMarshal(check pb.Checkpoint) ([]byte, error) {
	val := reflect.ValueOf(check)
	typ := val.Type()
	buf := make([]byte, 40)
	position := uint64(0)
	var err error
	for i := 0; i < typ.NumField(); i++ {
		if strings.Contains(typ.Field(i).Name, "XXX_") {
			continue
		}
		ff := typesMap[val.Field(i).Type()]
		position, err = ff(val.Field(i), buf, position)
		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

type structHasher struct {
	lock sync.Mutex
	hashCache *ccache.Cache
	cachedLeaves [][]byte
	layers [][][]byte
}

func (s *structHasher) fastHashTreeRoot(bs BeaconState) ([32]byte, error) {
	numItems := len(bs.BlockRoots)
	leaves := make([][]byte, numItems)
	hashKeyElements := make([]byte, BytesPerChunk*numItems)
	emptyKey := highwayhash.Sum(hashKeyElements, fastSumHashKey[:])
	offset := 0
	changedIndices := make([]int, 0)
	for i := 0; i < numItems; i++ {
		// If we are marshaling an byte array of length 32, we shortcut the computations and
		// simply return it as an identity root.
		leaves[i] = bs.BlockRoots[i][:]
		copy(hashKeyElements[offset:offset+32], leaves[i])
		offset += 32
		if s.cachedLeaves != nil {
			if !bytes.Equal(leaves[i], s.cachedLeaves[i]) {
				changedIndices = append(changedIndices, i)
			}
		}
	}
	hashKey := highwayhash.Sum(hashKeyElements, fastSumHashKey[:])
	if hashKey != emptyKey {
		s.lock.Lock()
		res := s.hashCache.Get(string(hashKey[:]))
		s.lock.Unlock()
		if res != nil && res.Value() != nil {
			return res.Value().([32]byte), nil
		}
	}
	chunks := leaves
	if len(changedIndices) > 0 {
		return s.recomputeBranch(changedIndices, chunks),  nil
	}
	root := s.merkleize(chunks)
	s.cachedLeaves = leaves
	if hashKey != emptyKey {
		s.lock.Lock()
		s.hashCache.Set(string(hashKey[:]), root, time.Hour)
		s.lock.Unlock()
	}
	return root, nil
}

func (s *structHasher) recomputeBranch(changedIndices []int, chunks [][]byte) [32]byte {
	left := s.layers[0][changedIndices[0]-1]
	right := chunks[changedIndices[0]]
	root := hash(append(left, right...))
	for i := 1; i < len(s.layers)-1; i++ {
		root = hash(append(s.layers[i][len(s.layers[i])-2], root[:]...))
	}
	return root
}

func (s *structHasher) merkleize(chunks [][]byte) [32]byte {
	if len(chunks) == 1 {
		var root [32]byte
		copy(root[:], chunks[0])
		return root
	}
	for !isPowerOf2(len(chunks)) {
		chunks = append(chunks, make([]byte, BytesPerChunk))
	}
	hashLayer := chunks
	s.layers[0] = hashLayer
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
		s.layers[i] = hashLayer
		i++
	}
	var root [32]byte
	copy(root[:], hashLayer[0])
	return root
}

func isPowerOf2(n int) bool {
	return n != 0 && (n&(n-1)) == 0
}

func marshalUint64(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	binary.LittleEndian.PutUint64(buf[startOffset:], val.Interface().(uint64))
	return startOffset + 8, nil
}

func marshalByteArray(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	copy(buf[startOffset:], val.Bytes())
	return startOffset + uint64(val.Len()), nil
}

// hash defines a function that returns the sha256 hash of the data passed in.
func hash(data []byte) [32]byte {
	return sha256.Sum256(data)
}

func toBytes32(x []byte) [32]byte {
	var y [32]byte
	copy(y[:], x)
	return y
}
