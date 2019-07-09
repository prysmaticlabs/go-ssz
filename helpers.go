package ssz

import (
	"bytes"
	"crypto/sha256"
	"math"
	"reflect"
)

var (
	// BytesPerChunk for an SSZ serialized object.
	BytesPerChunk = 32
	// BytesPerLengthOffset defines a constant for off-setting serialized chunks.
	BytesPerLengthOffset = uint64(4)
	zeroHashes           = make([][]byte, 32)
)

func init() {
	leaf := append([]byte{}, []byte{}...)
	result := hash(leaf)
	zeroHashes[0] = result[:]
	for i := 1; i < 32; i++ {
		leaf := append(zeroHashes[i-1], zeroHashes[i-1]...)
		result := hash(leaf)
		zeroHashes = append(zeroHashes, result[:])
	}
}

// Given ordered objects of the same basic type, serialize them, pack them into BYTES_PER_CHUNK-byte
// chunks, right-pad the last chunk with zero bytes, and return the chunks.
// Basic types are either bool, or uintN where N = {8, 16, 32, 64, 128, 256}.
//
// Important: due to limitations in Go generics, we will assume the input is already
// a list of SSZ-encoded objects of the same type.
func pack(serializedItems [][]byte) ([][]byte, error) {
	areAllEmpty := true
	for _, item := range serializedItems {
		if !bytes.Equal(item, []byte{}) {
			areAllEmpty = false
			break
		}
	}
	// If there are no items, we return an empty chunk.
	if len(serializedItems) == 0 || areAllEmpty {
		emptyChunk := make([]byte, BytesPerChunk)
		return [][]byte{emptyChunk}, nil
	} else if len(serializedItems[0]) == BytesPerChunk {
		// If each item has exactly BYTES_PER_CHUNK length, we return the list of serialized items.
		return serializedItems, nil
	}
	// We flatten the list in order to pack its items into byte chunks correctly.
	orderedItems := []byte{}
	for _, item := range serializedItems {
		orderedItems = append(orderedItems, item...)
	}
	numItems := len(orderedItems)
	chunks := [][]byte{}
	for i := 0; i < numItems; i += BytesPerChunk {
		j := i + BytesPerChunk
		// We create our upper bound index of the chunk, if it is greater than numItems,
		// we set it as numItems itself.
		if j > numItems {
			j = numItems
		}
		// We create chunks from the list of items based on the
		// indices determined above.
		chunks = append(chunks, orderedItems[i:j])
	}
	// Right-pad the last chunk with zero bytes if it does not
	// have length BytesPerChunk.
	lastChunk := chunks[len(chunks)-1]
	for len(lastChunk) < BytesPerChunk {
		lastChunk = append(lastChunk, 0)
	}
	chunks[len(chunks)-1] = lastChunk
	return chunks, nil
}

// Given ordered BYTES_PER_CHUNK-byte chunks, if necessary append zero chunks so that the
// number of chunks is a power of two, Merkleize the chunks, and return the root.
// Note that merkleize on a single chunk is simply that chunk, i.e. the identity
// when the number of chunks is one.
func merkleize(chunks [][]byte, hasPadding bool, padding uint64) [32]byte {
	if len(chunks) == 0 {
		zeroHash := make([]byte, 32)
		res := hash(append(zeroHash, zeroHash...))
		return res
	}
	if hasPadding {
		nextPowerOfTwo := padding
		for !isPowerTwo(int(nextPowerOfTwo)) {
			nextPowerOfTwo++
		}
		initialChunks := len(chunks)
		for i := uint64(initialChunks); i < nextPowerOfTwo; i++ {
			chunks = append(chunks, make([]byte, BytesPerChunk))
		}
	}

	for !isPowerTwo(len(chunks)) {
		chunks = append(chunks, make([]byte, BytesPerChunk))
	}
	if len(chunks) == 1 {
		var root [32]byte
		copy(root[:], chunks[0])
		return root
	}
	hashLayer := chunks
	// We keep track of the hash layers of a Merkle trie until we reach
	// the top layer of length 1, which contains the single root element.
	//        [Root]      -> Top layer has length 1.
	//    [E]       [F]   -> This layer has length 2.
	// [A]  [B]  [C]  [D] -> The bottom layer has length 4 (needs to be a power of two.
	for len(hashLayer) > 1 {
		layer := [][]byte{}
		for i := 0; i < len(hashLayer); i += 2 {
			hashedChunk := hash(append(hashLayer[i], hashLayer[i+1]...))
			layer = append(layer, hashedChunk[:])
		}
		hashLayer = layer
	}
	var root [32]byte
	copy(root[:], hashLayer[0])
	return root
}

func bitwiseMerkleize(chunks [][]byte, padding uint64) [32]byte {
	padTo := padding
	if padding == 0 {
		padTo = 1
	}
	count := uint64(len(chunks))
	depth := uint64(bitLength(0))
	if bitLength(count-1) > depth {
		depth = bitLength(count - 1)
	}
	maxDepth := depth
	if bitLength(padTo-1) > maxDepth {
		maxDepth = bitLength(padTo - 1)
	}
	layers := make([][]byte, maxDepth+1)

	for idx, chunk := range chunks {
		mergeChunks(layers, chunk, uint64(idx), count, depth)
	}

	if 1<<depth != count {
		mergeChunks(layers, zeroHashes[0], count, count, depth)
	}

	for i := depth; i < maxDepth; i++ {
		res := hash(append(layers[i], zeroHashes[i]...))
		layers[i+1] = res[:]
	}

	return toBytes32(layers[maxDepth])
}

func mergeChunks(layers [][]byte, currentRoot []byte, i, count, depth uint64) {
	j := uint64(0)
	for {
		if i&(1<<j) == 0 {
			if i == count && j < depth {
				res := hash(append(currentRoot[:], zeroHashes[j]...))
				currentRoot = res[:]
			} else {
				break
			}
		} else {
			res := hash(append(layers[j], currentRoot[:]...))
			currentRoot = res[:]
		}
		j++
	}
	layers[j] = currentRoot[:]
}

func bitLength(n uint64) uint64 {
	if n == 0 {
		return 1
	}
	return uint64(math.Log2(float64(n))) + 1
}

// Given a Merkle root root and a length length ("uint256" little-endian serialization)
// return hash(root + length).
func mixInLength(root [32]byte, length []byte) [32]byte {
	return hash(append(root[:], length...))
}

// Fast verification to check if an number if a power of two.
func isPowerTwo(n int) bool {
	return n != 0 && (n&(n-1)) == 0
}

// Instantiates a reflect value which may not have a concrete type to have a concrete type
// for unmarshaling. For example, we cannot unmarshal into a nil value - instead, it must have
// a concrete type even if all of its values are zero values.
func instantiateConcreteTypeForElement(val reflect.Value, typ reflect.Type) {
	val.Set(reflect.New(typ))
}

// Grows a slice to a new length and instantiates the element at length-1 with a concrete type
// accordingly if it is set to a pointer.
func growConcreteSliceType(val reflect.Value, typ reflect.Type, length int) {
	newVal := reflect.MakeSlice(typ, length, length)
	reflect.Copy(newVal, val)
	val.Set(newVal)
	if val.Index(length-1).Kind() == reflect.Ptr {
		instantiateConcreteTypeForElement(val.Index(length-1), typ.Elem().Elem())
	}
}

// toBytes32 is a convenience method for converting a byte slice to a fix
// sized 32 byte array. This method will truncate the input if it is larger
// than 32 bytes.
func toBytes32(x []byte) [32]byte {
	var y [32]byte
	copy(y[:], x)
	return y
}

// hash defines a function that returns the sha256 hash of the data passed in.
func hash(data []byte) [32]byte {
	var hash [32]byte

	h := sha256.New()
	// The hash interface never returns an error, for that reason
	// we are not handling the error below. For reference, it is
	// stated here https://golang.org/pkg/hash/#Hash
	// #nosec G104
	h.Write(data)
	h.Sum(hash[:0])

	return hash
}
