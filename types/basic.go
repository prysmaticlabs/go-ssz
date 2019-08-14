package types

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"sync"
)

type basicSSZ struct {
	hashCache map[string]interface{}
	lock      sync.Mutex
}

func newBasicSSZ() *basicSSZ {
	return &basicSSZ{
		hashCache: make(map[string]interface{}),
	}
}

func (b *basicSSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	kind := typ.Kind()
	switch {
	case kind == reflect.Bool:
		return marshalBool(val, buf, startOffset)
	case kind == reflect.Uint8:
		return marshalUint8(val, buf, startOffset)
	case kind == reflect.Uint16:
		return marshalUint16(val, buf, startOffset)
	case kind == reflect.Int32:
		return marshalInt32(val, buf, startOffset)
	case kind == reflect.Uint32:
		return marshalUint32(val, buf, startOffset)
	case kind == reflect.Uint64:
		return marshalUint64(val, buf, startOffset)
	case kind == reflect.Array && typ.Elem().Kind() == reflect.Uint8:
		return marshalByteArray(val, typ, buf, startOffset)
	case kind == reflect.Array && isBasicType(typ.Elem().Kind()):
		return b.marshalBasicArray(val, typ, buf, startOffset)
	default:
		return 0, fmt.Errorf("type %v is not serializable", val.Type())
	}
}

func (b *basicSSZ) Unmarshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	kind := typ.Kind()
	switch {
	case kind == reflect.Bool:
		return unmarshalBool(val, typ, buf, startOffset)
	case kind == reflect.Uint8:
		return unmarshalUint8(val, typ, buf, startOffset)
	case kind == reflect.Uint16:
		return unmarshalUint16(val, typ, buf, startOffset)
	case kind == reflect.Int32:
		return unmarshalInt32(val, typ, buf, startOffset)
	case kind == reflect.Uint32:
		return unmarshalUint32(val, typ, buf, startOffset)
	case kind == reflect.Uint64:
		return unmarshalUint64(val, typ, buf, startOffset)
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		return unmarshalByteArray(val, typ, buf, startOffset)
	case kind == reflect.Array && isBasicType(typ.Elem().Kind()):
		return basicArrayFactory.Unmarshal(val, typ, buf, startOffset)
	default:
		return 0, fmt.Errorf("type %v is not serializable", val.Type())
	}
}

func (b *basicSSZ) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	var chunks [][]byte
	var err error
	var hashKey string
	buf := make([]byte, DetermineSize(val))
	if _, err := b.Marshal(val, typ, buf, 0); err != nil {
		return [32]byte{}, err
	}
	hashKey = string(buf)
	b.lock.Lock()
	res := b.hashCache[hashKey]
	b.lock.Unlock()
	if res != nil {
		return res.([32]byte), nil
	}

	// In order to find the root of a basic type, we simply marshal it,
	// split the marshaling into chunks, and compute the most simple
	// Merkleization over the chunks.
	chunks, err = pack([][]byte{buf})
	if err != nil {
		return [32]byte{}, err
	}
	root, err := bitwiseMerkleize(chunks, uint64(len(chunks)), uint64(len(chunks)))
	if err != nil {
		return [32]byte{}, err
	}
	b.lock.Lock()
	b.hashCache[hashKey] = root
	b.lock.Unlock()
	return root, nil
}

func (b *basicSSZ) marshalBasicArray(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	index := startOffset
	var err error
	for i := 0; i < val.Len(); i++ {
		index, err = b.Marshal(val.Index(i), typ.Elem(), buf, index)
		if err != nil {
			return 0, err
		}
	}
	return index, nil
}

func marshalByteArray(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	if val.Type().Kind() == reflect.Array {
		for i := 0; i < val.Len(); i++ {
			buf[int(startOffset)+i] = uint8(val.Index(i).Uint())
		}
		return startOffset + uint64(val.Len()), nil
	}
	copy(buf[startOffset:], val.Bytes())
	return startOffset + uint64(val.Len()), nil
}

func unmarshalByteArray(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	offset := startOffset + uint64(len(input))
	val.SetBytes(input[startOffset:offset])
	return offset, nil
}

func marshalBool(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	if val.Interface().(bool) {
		buf[startOffset] = uint8(1)
	} else {
		buf[startOffset] = uint8(0)
	}
	return startOffset + 1, nil
}

func unmarshalBool(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	v := input[startOffset]
	if v == 0 {
		val.SetBool(false)
	} else if v == 1 {
		val.SetBool(true)
	} else {
		return 0, fmt.Errorf("expected 0 or 1 but received %d", v)
	}
	return startOffset + 1, nil
}

func marshalUint8(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	buf[startOffset] = val.Interface().(uint8)
	return startOffset + 1, nil
}

func unmarshalUint8(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	val.SetUint(uint64(input[startOffset]))
	return startOffset + 1, nil
}

func marshalUint16(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	binary.LittleEndian.PutUint16(buf[startOffset:], val.Interface().(uint16))
	return startOffset + 2, nil
}

func unmarshalUint16(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	offset := startOffset + 2
	buf := make([]byte, 2)
	copy(buf, input[startOffset:offset])
	val.SetUint(uint64(binary.LittleEndian.Uint16(buf)))
	return offset, nil
}

func marshalInt32(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	binary.LittleEndian.PutUint32(buf[startOffset:], uint32(val.Interface().(int32)))
	return startOffset + 4, nil
}

func unmarshalInt32(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	offset := startOffset + 4
	buf := make([]byte, 4)
	copy(buf, input[startOffset:offset])
	val.SetInt(int64(binary.LittleEndian.Uint32(buf)))
	return offset, nil
}

func marshalUint32(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	binary.LittleEndian.PutUint32(buf[startOffset:], val.Interface().(uint32))
	return startOffset + 4, nil
}

func unmarshalUint32(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	offset := startOffset + 4
	buf := make([]byte, 4)
	copy(buf, input[startOffset:offset])
	val.SetUint(uint64(binary.LittleEndian.Uint32(buf)))
	return offset, nil
}

func marshalUint64(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	binary.LittleEndian.PutUint64(buf[startOffset:], val.Interface().(uint64))
	return startOffset + 8, nil
}

func unmarshalUint64(val reflect.Value, typ reflect.Type, input []byte, startOffset uint64) (uint64, error) {
	offset := startOffset + 8
	buf := make([]byte, 8)
	copy(buf, input[startOffset:offset])
	val.SetUint(binary.LittleEndian.Uint64(buf))
	return offset, nil
}
