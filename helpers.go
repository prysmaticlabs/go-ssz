package ssz

import (
	"bytes"
	"github.com/prysmaticlabs/prysm/shared/ssz"
)

// Given ordered objects of the same basic type, serialize them, pack them into BYTES_PER_CHUNK-byte
// chunks, right-pad the last chunk with zero bytes, and return the chunks.
// Basic types are either bool, or uintN where N = {8, 16, 32, 64, 128, 256}.
func pack(objects []interface{}) ([][]byte, error) {
    serializedItems := make([][]byte, len(objects))
    for i, item := range objects {
		// We use a bytes.Buffer as our io.Writer.
		buffer := new(bytes.Buffer)
		// ssz.Encode writes the encoded data to the buffer.
		if err := ssz.Encode(buffer, item); err != nil {
			return nil, err
		}
		serializedItems[i] = buffer.Bytes()
	}
	chunks := [][]byte{}
	numItems := len(serializedItems)
	if numItems == 0 {
		emptyChunk := make([]byte, BytesPerChunk)
		return [][]byte{emptyChunk}, nil
	} else if len(serializedItems[0]) == BytesPerChunk {
		return serializedItems, nil
	}
	itemsPerChunk := sszChunkSize / len(serializedItems[0])
	for i := 0; i < numItems; i += itemsPerChunk {
		chunk := make([]byte, 0)
		j := i + itemsPerChunk
		if j > numItems {
			j = numItems
		}
		for _, item := range serializedItems[i:j] {
			chunk = append(chunk, item...)
		}
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}

// Given ordered BYTES_PER_CHUNK-byte chunks, if necessary append zero chunks so that the
// number of chunks is a power of two, Merkleize the chunks, and return the root.
// Note that merkleize on a single chunk is simply that chunk, i.e. the identity
// when the number of chunks is one.
func merkleize(chunks [][]byte) ([32]byte, error) {
	return [32]byte{}, nil
}

// Given a Merkle root root and a length length ("uint256" little-endian serialization)
// return hash(root + length).
func mixInLength(root [32]byte, length []byte) [32]byte {
	return Hash(append(root[:], length...))
}

// Given a Merkle root root and a type_index type_index ("uint256" little-endian serialization)
// return hash(root + type_index).
func mixInType(root [32]byte, typeIndex []byte) [32]byte {
	return Hash(append(root[:], typeIndex...))
}
