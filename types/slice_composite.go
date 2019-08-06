package types

import (
	"bytes"
	"encoding/binary"
	"reflect"
)

type CompositeSliceSSZ struct {
	hashCache map[string]interface{}
}

func NewCompositeSliceSSZ() *CompositeSliceSSZ {
	return &CompositeSliceSSZ{
		hashCache: make(map[string]interface{}),
	}
}

func (b *CompositeSliceSSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	output := make([]byte, 32)
	if val.Len() == 0 && maxCapacity == 0 {
		merkleRoot, err := bitwiseMerkleize([][]byte{}, 0, 0)
		if err != nil {
			return [32]byte{}, err
		}
		return mixInLength(merkleRoot, output), nil
	}
	numItems := val.Len()
	var factory SSZAble
	var err error
	if numItems > 0 {
		factory, err = SSZFactory(val.Index(0), typ.Elem())
		if err != nil {
			return [32]byte{}, err
		}
	}
	roots := make([][]byte, numItems)
	hashKey := make([]byte, 32*numItems)
	offset := 0
	for i := 0; i < numItems; i++ {
		r, err := factory.Root(val.Index(i), typ.Elem(), 0)
		if err != nil {
			return [32]byte{}, err
		}
		roots[i] = r[:]
		copy(hashKey[offset:offset+32], r[:])
		offset += 32
	}
	chunks, err := pack(roots)
	if err != nil {
		return [32]byte{}, err
	}
	if numItems > 0 {
		res := b.hashCache[string(hashKey)]
		if res != nil {
			return res.([32]byte), nil
		}
	}
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, uint64(val.Len())); err != nil {
		return [32]byte{}, err
	}
	copy(output, buf.Bytes())
	objLen := maxCapacity
	if maxCapacity == 0 {
		objLen = uint64(val.Len())
	}
	merkleRoot, err := bitwiseMerkleize(chunks, uint64(len(chunks)), objLen)
	if err != nil {
		return [32]byte{}, err
	}
	result := mixInLength(merkleRoot, output)
	b.hashCache[string(hashKey)] = result
	return result, nil
}

func (b *CompositeSliceSSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	index := startOffset
	if val.Len() == 0 {
		return index, nil
	}
	factory, err := SSZFactory(val.Index(0), typ.Elem())
	if err != nil {
		return 0, err
	}
	if !isVariableSizeType(typ.Elem()) {
		for i := 0; i < val.Len(); i++ {
			// If each element is not variable size, we simply encode sequentially and write
			// into the buffer at the last index we wrote at.
			index, err = factory.Marshal(val.Index(i), typ.Elem(), buf, index)
			if err != nil {
				return 0, err
			}
		}
		return index, nil
	}
	fixedIndex := index
	currentOffsetIndex := startOffset + uint64(val.Len())*BytesPerLengthOffset
	nextOffsetIndex := currentOffsetIndex
	// If the elements are variable size, we need to include offset indices
	// in the serialized output list.
	for i := 0; i < val.Len(); i++ {
		nextOffsetIndex, err = factory.Marshal(val.Index(i), typ.Elem(), buf, currentOffsetIndex)
		if err != nil {
			return 0, err
		}
		// Write the offset.
		offsetBuf := make([]byte, BytesPerLengthOffset)
		binary.LittleEndian.PutUint32(offsetBuf, uint32(currentOffsetIndex-startOffset))
		copy(buf[fixedIndex:fixedIndex+BytesPerLengthOffset], offsetBuf)

		// We increase the offset indices accordingly.
		currentOffsetIndex = nextOffsetIndex
		fixedIndex += BytesPerLengthOffset
	}
	index = currentOffsetIndex
	return index, nil
}

func (b *CompositeSliceSSZ) Unmarshal(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	if len(input) == 0 {
		newVal := reflect.MakeSlice(val.Type(), 0, 0)
		val.Set(newVal)
		return 0, nil
	}
	growConcreteSliceType(val, typ, 1)
	endOffset := uint64(len(input))

	currentIndex := startOffset
	nextIndex := currentIndex
	offsetVal := input[startOffset : startOffset+BytesPerLengthOffset]
	firstOffset := startOffset + uint64(binary.LittleEndian.Uint32(offsetVal))
	currentOffset := firstOffset
	nextOffset := currentOffset
	i := 0
	for currentIndex < firstOffset {
		nextIndex = currentIndex + BytesPerLengthOffset
		if nextIndex == firstOffset {
			nextOffset = endOffset
		} else {
			nextOffsetVal := input[nextIndex : nextIndex+BytesPerLengthOffset]
			nextOffset = startOffset + uint64(binary.LittleEndian.Uint32(nextOffsetVal))
		}
		// We grow the slice's size to accommodate a new element being unmarshaled.
		growConcreteSliceType(val, typ, i+1)
		factory, err := SSZFactory(val.Index(i), typ.Elem())
		if err != nil {
			return 0, err
		}
		if _, err := factory.Unmarshal(val.Index(i), typ.Elem(), input[currentOffset:nextOffset], 0); err != nil {
			return 0, err
		}
		i++
		currentIndex = nextIndex
		currentOffset = nextOffset
	}
	return currentIndex, nil
}
