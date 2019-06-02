package ssz

// Given ordered objects of the same basic type, serialize them, pack them into BYTES_PER_CHUNK-byte
// chunks, right-pad the last chunk with zero bytes, and return the chunks.
func pack(vals []interface) ([][]byte, error) {
	return [][]byte{}, nil
}

// Given ordered BYTES_PER_CHUNK-byte chunks, if necessary append zero chunks so that the
// number of chunks is a power of two, Merkleize the chunks, and return the root.
// Note that merkleize on a single chunk is simply that chunk, i.e. the identity
// when the number of chunks is one.
func merkleize(vals []interface) ([32]byte, error) {
	return [32]byte{}, nil
}

// Given a Merkle root root and a length length ("uint256" little-endian serialization)
// return hash(root + length).
func mixInLength(root [32]byte, length uint64) ([32]byte, error) {
	return [32]byte{}, nil
}

// Given a Merkle root root and a type_index type_index ("uint256" little-endian serialization)
// return hash(root + type_index).
func mixInType(root [32]byte, typeIndex uint64) ([32]byte, error) {
	return [32]byte{}, nil
}
