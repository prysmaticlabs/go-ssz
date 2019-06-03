package ssz

import (
	"bytes"
	"github.com/prysmaticlabs/prysm/shared/ssz"
)

// Given ordered objects of the same basic type, serialize them, pack them into BYTES_PER_CHUNK-byte
// chunks, right-pad the last chunk with zero bytes, and return the chunks.
func pack(objects []interface{}) ([][]byte, error) {
	// We use a bytes.Buffer as our io.Writer.
	buffer := new(bytes.Buffer)
	// ssz.Encode writes the encoded data to the buffer.
	if err := ssz.Encode(buffer, objects); err != nil {
		return nil, err
	}
	encodedBytes := buffer.Bytes()
	chunks := [][]byte{}
	if len(encodedBytes) == 0 {
		emptyChunk := make([]byte, BytesPerChunk)
		return [][]byte{emptyChunk}, nil
	} else if len(encodedBytes) == BytesPerChunk {
		return [][]byte{encodedBytes}, nil
	}
	// Otherwise, we pack items into chunks and pad if necessary.
	return [][]byte{}, nil
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
