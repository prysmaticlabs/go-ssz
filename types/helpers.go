package types

import (
	"bytes"
	"errors"
	"reflect"

	"github.com/minio/sha256-simd"
	"github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/merkle"
)

var (
	// BytesPerChunk for an SSZ serialized object.
	BytesPerChunk = 32
	// BytesPerLengthOffset defines a constant for off-setting serialized chunks.
	BytesPerLengthOffset = uint64(4)
	zeroHashes           = make([][32]byte, 100)
)

func init() {
	for i := 1; i < 100; i++ {
		leaf := append(zeroHashes[i-1][:], zeroHashes[i-1][:]...)
		result := hash(leaf)
		copy(zeroHashes[i][:], result[:])
	}
}

// Given ordered BYTES_PER_CHUNK-byte chunks, if necessary utilize zero chunks so that the
// number of chunks is a power of two, Merkleize the chunks, and return the root.
// Note that merkleize on a single chunk is simply that chunk, i.e. the identity
// when the number of chunks is one.
func bitwiseMerkleize(chunks [][]byte, count uint64, limit uint64) ([32]byte, error) {
	if count > limit {
		return [32]byte{}, errors.New("merkleizing list that is too large, over limit")
	}
	hasher := htr.HashFn(hash)
	leafIndexer := func(i uint64) []byte {
		return chunks[i]
	}
	return merkle.Merkleize(hasher, count, limit, leafIndexer), nil
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

// Given a Merkle root root and a length length ("uint256" little-endian serialization)
// return hash(root + length).
func mixInLength(root [32]byte, length []byte) [32]byte {
	var hash [32]byte
	h := sha256.New()
	h.Write(root[:])
	h.Write(length)
	// The hash interface never returns an error, for that reason
	// we are not handling the error below. For reference, it is
	// stated here https://golang.org/pkg/hash/#Hash
	// #nosec G104
	h.Sum(hash[:0])
	return hash
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

// hash defines a function that returns the sha256 hash of the data passed in.
func hash(data []byte) [32]byte {
	return sha256.Sum256(data)
}

func growSliceFromSizeTags(val reflect.Value, sizes []uint64) reflect.Value {
	if len(sizes) == 0 {
		return val
	}
	finalValue := reflect.MakeSlice(val.Type(), int(sizes[0]), int(sizes[0]))
	for i := 0; i < int(sizes[0]); i++ {
		intermediate := growSliceFromSizeTags(finalValue.Index(i), sizes[1:])
		finalValue.Index(i).Set(intermediate)
	}
	return finalValue
}
