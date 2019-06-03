package ssz

import (
	"errors"
)

var (
	BytesPerChunk = 32 // BytesPerChunk for an SSZ serialized object.
	BytesPerLengthOffset = 4 // BytesPerLengthOffset defines a constant for off-setting serialized chunks.
	BitsPerByte = 8 // BitsPerByte as a useful constant.
	ErrNotMerkleRoot = errors.New("object is not a Merkle root") // ErrNotMerkleRoot when a cached item is not root.
)
