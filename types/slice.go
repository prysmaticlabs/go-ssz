package types

import (
	"bytes"
	"encoding/binary"
	"reflect"
)

type BasicSliceSSZ struct {
	hashCache map[string]interface{}
}

type CompositeSliceSSZ struct {
	hashCache map[string]interface{}
}

type BasicArraySSZ struct {
	hashCache map[string]interface{}
}

type CompositeArraySSZ struct {
	hashCache map[string]interface{}
}

type BitlistSSZ struct {
	hashCache map[string]interface{}
}

func NewBasicSliceSSZ() *BasicSliceSSZ {
	return &BasicSliceSSZ{
		hashCache: make(map[string]interface{}),
	}
}

func NewBasicArraySSZ() *BasicArraySSZ {
	return &BasicArraySSZ{
		hashCache: make(map[string]interface{}),
	}
}

func NewCompositeArraySSZ() *CompositeArraySSZ {
	return &CompositeArraySSZ{
		hashCache: make(map[string]interface{}),
	}
}

func NewCompositeSliceSSZ() *CompositeSliceSSZ {
	return &CompositeSliceSSZ{
		hashCache: make(map[string]interface{}),
	}
}

func (b *CompositeSliceSSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	output := make([]byte, 32)
	if val.Len() == 0 && maxCapacity == 0 {
		//merkleRoot, err := bitwiseMerkleize([][]byte{}, 0, true /* has limit */)
		merkleRoot, err := NewMerkleize([][]byte{}, 0, 0, true)
		if err != nil {
			return [32]byte{}, err
		}
		return mixInLength(merkleRoot, output), nil
	}
	factory, _ := SSZFactory(val.Index(0), typ.Elem())
	numItems := val.Len()
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

	res := b.hashCache[string(hashKey)]
	if res != nil {
		return res.([32]byte), nil
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
	merkleRoot, err := NewMerkleize(roots, uint64(len(roots)), objLen, true)
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

func (b *BasicSliceSSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	var factory SSZAble
	elemSize := determineFixedSize(val, typ.Elem())
	limit := (maxCapacity*elemSize + 31) / 32
	numItems := val.Len()
	if numItems > 0 {
		factory, _ = SSZFactory(val.Index(0), typ.Elem())
		// TODO: Abstract into a chunk count function.
	}
	if limit == 0 {
		limit = 1
	}
	leaves := make([][]byte, numItems)
	var err error
	for i := 0; i < numItems; i++ {
		innerBuf := make([]byte, elemSize)
		if _, err = factory.Marshal(val.Index(i), typ.Elem(), innerBuf, 0); err != nil {
			return [32]byte{}, err
		}
		leaves[i] = innerBuf
	}
	chunks, err := pack(leaves)
	if err != nil {
		return [32]byte{}, err
	}
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, uint64(val.Len())); err != nil {
		return [32]byte{}, err
	}
	output := make([]byte, 32)
	copy(output, buf.Bytes())
	merkleRoot, err := NewMerkleize(chunks, uint64(len(chunks)), limit, true)
	if err != nil {
		return [32]byte{}, err
	}
	return mixInLength(merkleRoot, output), nil
}

func (b *BasicSliceSSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	index := startOffset
	var err error
	if val.Len() == 0 {
		return index, nil
	}
	factory, err := SSZFactory(val.Index(0), typ.Elem())
	if err != nil {
		return 0, err
	}
	for i := 0; i < val.Len(); i++ {
		index, err = factory.Marshal(val.Index(i), typ.Elem(), buf, index)
		if err != nil {
			return 0, err
		}
	}
	return index, nil
}

func (b *CompositeArraySSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	factory, _ := SSZFactory(val.Index(0), typ.Elem())
	numItems := val.Len()
	roots := make([][]byte, numItems)
	hashKey := make([]byte, 32*numItems)
	limit := uint64(typ.Len())
	offset := 0
	var err error
	for i := 0; i < val.Len(); i++ {
		r, err := factory.Root(val.Index(i), typ.Elem(), 0)
		if err != nil {
			return [32]byte{}, err
		}
		roots[i] = r[:]
		copy(hashKey[offset:offset+32], r[:])
		offset += 32
	}

	res := b.hashCache[string(hashKey)]
	if res != nil {
		return res.([32]byte), nil
	}
	//root, err := bitwiseMerkleize(roots, limit, true /* has limit */)
	root, err := NewMerkleize(roots, uint64(len(roots)), limit, true)
	if err != nil {
		return [32]byte{}, err
	}
	b.hashCache[string(hashKey)] = root
	return root, nil
}

func (b *CompositeArraySSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
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

func (b *BasicArraySSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	numItems := val.Len()
	hashKey := make([]byte, 32*numItems)
	leaves := make([][]byte, numItems)
	elemKind := typ.Elem().Kind()
	offset := 0
	var err error
	for i := 0; i < numItems; i++ {
		if elemKind == reflect.Array && val.Index(i).Len() == 32 {
			leaves[i] = val.Index(i).Bytes()
			copy(hashKey[offset:offset+32], leaves[i])
			continue
		}
		r, err := BasicFactory.Root(val.Index(i), typ.Elem(), 0)
		if err != nil {
			return [32]byte{}, err
		}
		leaves[i] = r[:]
		copy(hashKey[offset:offset+32], r[:])
	}
	res := b.hashCache[string(hashKey)]
	if res != nil {
		return res.([32]byte), nil
	}
	root, err := NewMerkleize(leaves, uint64(len(leaves)), uint64(len(leaves)), false /* has limit */)
	if err != nil {
		return [32]byte{}, err
	}
	b.hashCache[string(hashKey)] = root
	return root, nil
}

func (b *BasicArraySSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	index := startOffset
	var err error
	if val.Len() == 0 {
		return index, nil
	}
	factory, err := SSZFactory(val.Index(0), typ.Elem())
	if err != nil {
		return 0, err
	}
	for i := 0; i < val.Len(); i++ {
		index, err = factory.Marshal(val.Index(i), typ.Elem(), buf, index)
		if err != nil {
			return 0, err
		}
	}
	return index, nil
}
