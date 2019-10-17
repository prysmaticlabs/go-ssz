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
	"github.com/minio/sha256-simd"

	"github.com/prysmaticlabs/go-ssz"
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
	log.Println("---Starting regular HTR...")
	start := time.Now()
	enc, err := ssz.HashTreeRoot(bs)
	if err != nil {
		panic(err)
	}
	end := time.Now()
	log.Printf("%#x in %v", enc, end.Sub(start))
	log.Println("---Ended regular HTR")
	start = time.Now()
	res, err := hasher.fastHashTreeRoot(bs)
	if err != nil {
		panic(err)
	}
	end = time.Now()
	log.Printf("%#x in %v", res, end.Sub(start))
	// Change 2 elements.
	bs.BlockRoots[len(bs.BlockRoots)-1] = [32]byte{4, 5, 6}
	bs.BlockRoots[1] = [32]byte{7, 8, 9}
	start = time.Now()
	res, err = hasher.fastHashTreeRoot(bs)
	if err != nil {
		panic(err)
	}
	end = time.Now()
	log.Printf("%#x in %v", res, end.Sub(start))
	log.Println("---Starting regular HTR...")
	start = time.Now()
	enc, err = ssz.HashTreeRoot(bs)
	if err != nil {
		panic(err)
	}
	end = time.Now()
	log.Printf("%#x in %v", enc, end.Sub(start))
	log.Println("---Ended regular HTR")
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
	changedIndices := make([]int, 0)
	for i := 0; i < numItems; i++ {
		// If we are marshaling an byte array of length 32, we shortcut the computations and
		// simply return it as an identity root.
		leaves[i] = bs.BlockRoots[i][:]
		if s.cachedLeaves != nil {
			if !bytes.Equal(leaves[i], s.cachedLeaves[i]) {
				changedIndices = append(changedIndices, i)
			}
		}
	}
	chunks := leaves
	if len(changedIndices) > 0 {
		log.Printf("%d items changed", len(changedIndices))
		var rt [32]byte
		for i := 0; i < len(changedIndices); i++ {
			rt = s.recomputeRoot(changedIndices[i], chunks)
		}
		return rt, nil
	}
	root := s.merkleize(chunks)
	s.cachedLeaves = leaves
	return root, nil
}

func (s *structHasher) recomputeRoot(idx int, chunks [][]byte) [32]byte {
	root := chunks[idx]
	for i := 0; i < len(s.layers)-1; i++ {
		subIndex := (uint64(idx) / (1 << uint64(i))) ^ 1
		isLeft := uint64(idx) / (1 << uint64(i))
		parentIdx := uint64(idx) / (1 << uint64(i+1))
		item := s.layers[i][subIndex]
		if isLeft%2 != 0 {
			parentHash := hash(append(item, root...))
			root = parentHash[:]
		} else {
			parentHash := hash(append(root, item...))
			root = parentHash[:]
		}
		s.layers[i+1][parentIdx] = root
	}
	return toBytes32(root)
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
