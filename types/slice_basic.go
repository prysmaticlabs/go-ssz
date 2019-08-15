package types

import (
	"bytes"
	"encoding/binary"
	"reflect"
)

type basicSliceSSZ struct {
	hashCache map[string]interface{}
}

func newBasicSliceSSZ() *basicSliceSSZ {
	return &basicSliceSSZ{
		hashCache: make(map[string]interface{}),
	}
}

func (b *basicSliceSSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	var factory SSZAble
	var limit uint64
	var elemSize uint64
	var err error
	numItems := val.Len()
	if numItems > 0 {
		factory, err = SSZFactory(val.Index(0), typ.Elem())
		if err != nil {
			return [32]byte{}, err
		}
	}

	if isBasicType(typ.Elem().Kind()) {
		elemSize = determineFixedSize(val, typ.Elem())
	} else {
		elemSize = 32
	}
	limit = (maxCapacity*elemSize + 31) / 32
	if limit == 0 {
		if numItems == 0 {
			limit = 1
		} else {
			limit = uint64(numItems)
		}
	}
	leaves := make([][]byte, numItems)
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
			leaves[i] = r[:]
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

func (b *basicSliceSSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
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

func (b *basicSliceSSZ) Unmarshal(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
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
