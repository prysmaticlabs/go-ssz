package types

import (
	"bytes"
	"encoding/binary"
	"reflect"
)

type stringSSZ struct{}

func newStringSSZ() *stringSSZ {
	return &stringSSZ{}
}

func (b *stringSSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	var err error
	numItems := val.Len()
	elemSize := uint64(1)
	limit := (maxCapacity*elemSize + 31) / 32
	if limit == 0 {
		limit = 1
	}
	leaves := make([][]byte, numItems)
	for i := 0; i < numItems; i++ {
		innerBuf := make([]byte, elemSize)
		if _, err = marshalUint8(val.Index(i), innerBuf, 0); err != nil {
			return [32]byte{}, err
		}
		leaves[i] = innerBuf
	}
	chunks, err := pack(leaves)
	if err != nil {
		return [32]byte{}, err
	}
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, uint64(numItems)); err != nil {
		return [32]byte{}, err
	}
	output := make([]byte, 32)
	copy(output, buf.Bytes())
	merkleRoot, err := bitwiseMerkleize(chunks, uint64(len(chunks)), limit)
	if err != nil {
		return [32]byte{}, err
	}
	return mixInLength(merkleRoot, output), nil
}

func (b *stringSSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	for i := 0; i < val.Len(); i++ {
		buf[int(startOffset)+i] = uint8(val.Index(i).Uint())
	}
	return startOffset + uint64(val.Len()), nil
}

func (b *stringSSZ) Unmarshal(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	offset := startOffset + uint64(len(input))
	val.SetString(string(input[startOffset:offset]))
	return offset, nil
}
