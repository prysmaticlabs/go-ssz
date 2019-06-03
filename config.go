package ssz

import (
	"errors"
)

var (
	// BytesPerChunk for an SSZ serialized object.
	BytesPerChunk = 32
	// BytesPerLengthOffset defines a constant for off-setting serialized chunks.
	BytesPerLengthOffset = 4
	// BitsPerByte as a useful constant.
	BitsPerByte = 8
	// ErrNotMerkleRoot when a cached item is not root.
	ErrNotMerkleRoot = errors.New("object is not a Merkle root")
)
