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

func (b *BasicSliceSSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	var factory SSZAble
	var limit uint64
	var elemSize uint64
	numItems := val.Len()
	if numItems > 0 {
		factory, _ = SSZFactory(val.Index(0), typ.Elem())
		// TODO: Abstract into a chunk count function.
	}
	if isBasicType(typ.Elem().Kind()) {
		elemSize = determineFixedSize(val, typ.Elem())
	} else {
		elemSize = 32
	}
	limit = (maxCapacity*elemSize + 31) / 32
	if limit == 0 {
		limit = 1
	}
	leaves := make([][]byte, numItems)
	var err error
	for i := 0; i < numItems; i++ {
		if isBasicType(val.Index(i).Kind()) {
			innerBuf := make([]byte, elemSize)
			if _, err = factory.Marshal(val.Index(i), typ.Elem(), innerBuf, 0); err != nil {
				return [32]byte{}, err
			}
			leaves[i] = innerBuf
		} else {
			r, err := factory.Root(val.Index(i), typ.Elem(), 0)
			if err != nil {
				return [32]byte{}, err
			}
			leaves = append(leaves, r[:])
		}
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
	merkleRoot, err := bitwiseMerkleize(chunks, uint64(len(chunks)), limit)
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

func (b *BasicSliceSSZ) Unmarshal(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	if len(input) == 0 {
		newVal := reflect.MakeSlice(val.Type(), 0, 0)
		val.Set(newVal)
		return 0, nil
	}
	// If there are struct tags that specify a different type, we handle accordingly.
	if val.Type() != typ {
		sizes := []uint64{1}
		innerElement := typ.Elem()
		for {
			if innerElement.Kind() == reflect.Slice {
				sizes = append(sizes, 0)
				innerElement = innerElement.Elem()
			} else if innerElement.Kind() == reflect.Array {
				sizes = append(sizes, uint64(innerElement.Len()))
				innerElement = innerElement.Elem()
			} else {
				break
			}
		}
		// If the item is a slice, we grow it accordingly based on the size tags.
		result := growSliceFromSizeTags(val, sizes)
		reflect.Copy(result, val)
		val.Set(result)
	} else {
		growConcreteSliceType(val, val.Type(), 1)
	}

	var err error
	index := startOffset
	factory, err := SSZFactory(val.Index(0), typ.Elem())
	if err != nil {
		return 0, err
	}
	index, err = factory.Unmarshal(val.Index(0), typ.Elem(), input, index)
	if err != nil {
		return 0, err
	}

	elementSize := index - startOffset
	endOffset := uint64(len(input)) / elementSize
	if val.Type() != typ {
		sizes := []uint64{endOffset}
		innerElement := typ.Elem()
		for {
			if innerElement.Kind() == reflect.Slice {
				sizes = append(sizes, 0)
				innerElement = innerElement.Elem()
			} else if innerElement.Kind() == reflect.Array {
				sizes = append(sizes, uint64(innerElement.Len()))
				innerElement = innerElement.Elem()
			} else {
				break
			}
		}
		// If the item is a slice, we grow it accordingly based on the size tags.
		result := growSliceFromSizeTags(val, sizes)
		reflect.Copy(result, val)
		val.Set(result)
	}
	i := uint64(1)
	for i < endOffset {
		if val.Type() == typ {
			growConcreteSliceType(val, val.Type(), int(i)+1)
		}
		index, err = factory.Unmarshal(val.Index(int(i)), typ.Elem(), input, index)
		if err != nil {
			return 0, err
		}
		i++
	}
	return index, nil
}

func (b *CompositeArraySSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	factory, err := SSZFactory(val.Index(0), typ.Elem())
	if err != nil {
		return [32]byte{}, err
	}
	numItems := val.Len()
	roots := make([][]byte, numItems)
	hashKey := make([]byte, 32*numItems)
	elemSize := uint64(0)
	if isBasicType(typ.Elem().Kind()) {
		elemSize = determineFixedSize(val, typ.Elem())
	} else {
		elemSize = 32
	}
	limit := (uint64(val.Len())*elemSize + 31) / 32
	offset := 0
	for i := 0; i < val.Len(); i++ {
		r, err := factory.Root(val.Index(i), typ.Elem(), 0)
		if err != nil {
			return [32]byte{}, err
		}
		roots[i] = r[:]
		copy(hashKey[offset:offset+32], r[:])
		offset += 32
	}
	if val.Len() > 0 {
		res := b.hashCache[string(hashKey)]
		if res != nil {
			return res.([32]byte), nil
		}
	}
	chunks, err := pack(roots)
	if err != nil {
		return [32]byte{}, err
	}
	if val.Len() == 0 {
		chunks = [][]byte{}
	}
	root, err := bitwiseMerkleize(chunks, uint64(len(chunks)), limit)
	if err != nil {
		return [32]byte{}, err
	}
	if val.Len() > 0 {
		b.hashCache[string(hashKey)] = root
	}
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

func (b *CompositeArraySSZ) Unmarshal(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	currentIndex := startOffset
	nextIndex := currentIndex
	offsetVal := input[startOffset : startOffset+BytesPerLengthOffset]
	firstOffset := startOffset + uint64(binary.LittleEndian.Uint32(offsetVal))
	currentOffset := firstOffset
	nextOffset := currentOffset
	endOffset := uint64(len(input))
	i := 0
	factory, err := SSZFactory(val.Index(0), typ.Elem())
	if err != nil {
		return 0, err
	}
	for currentIndex < firstOffset {
		nextIndex = currentIndex + BytesPerLengthOffset
		if nextIndex == firstOffset {
			nextOffset = endOffset
		} else {
			nextOffsetVal := input[nextIndex : nextIndex+BytesPerLengthOffset]
			nextOffset = startOffset + uint64(binary.LittleEndian.Uint32(nextOffsetVal))
		}
		if val.Index(i).Kind() == reflect.Ptr {
			instantiateConcreteTypeForElement(val.Index(i), typ.Elem().Elem())
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

func (b *BasicArraySSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	numItems := val.Len()
	hashKey := make([]byte, 32*numItems)
	leaves := make([][]byte, numItems)
	elemKind := typ.Elem().Kind()
	offset := 0
	factory, err := SSZFactory(val.Index(0), typ.Elem())
	if err != nil {
		return [32]byte{}, err
	}
	for i := 0; i < numItems; i++ {
		if elemKind == reflect.Array && val.Index(i).Len() == 32 {
			leaves[i] = val.Index(i).Bytes()
			copy(hashKey[offset:offset+32], leaves[i])
			continue
		}
		r, err := factory.Root(val.Index(i), typ.Elem(), 0)
		if err != nil {
			return [32]byte{}, err
		}
		leaves[i] = r[:]
		copy(hashKey[offset:offset+32], r[:])
	}
	//if numItems > 0 {
	//	res := b.hashCache[string(hashKey)]
	//	if res != nil {
	//		return res.([32]byte), nil
	//	}
	//}
	chunks, err := pack(leaves)
	if err != nil {
		return [32]byte{}, err
	}
	root, err := bitwiseMerkleize(chunks, uint64(len(chunks)), uint64(len(chunks)))
	if err != nil {
		return [32]byte{}, err
	}
	//if numItems > 0 {
	//	b.hashCache[string(hashKey)] = root
	//}
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

func (b *BasicArraySSZ) Unmarshal(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	i := 0
	index := startOffset
	size := val.Len()
	var err error
	var factory SSZAble
	for i < size {
		if val.Index(i).Kind() == reflect.Ptr {
			instantiateConcreteTypeForElement(val.Index(i), typ.Elem().Elem())
			factory, err = SSZFactory(val.Index(i), typ.Elem().Elem())
			if err != nil {
				return 0, err
			}
		} else {
			factory, err = SSZFactory(val.Index(i), typ.Elem())
			if err != nil {
				return 0, err
			}
		}
		index, err = factory.Unmarshal(val.Index(i), typ.Elem(), input, index)
		if err != nil {
			return 0, err
		}
		i++
	}
	return index, nil
}
