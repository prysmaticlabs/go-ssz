package types

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"sync"

	"github.com/dgraph-io/ristretto"
)

// BasicTypeCacheSize for HashTreeRoot.
const BasicTypeCacheSize = 100000

type basicSSZ struct {
	hashCache *ristretto.Cache
	lock      sync.Mutex
}

func newBasicSSZ() *basicSSZ {
	cache, _ := ristretto.NewCache(&ristretto.Config{
		NumCounters: BasicTypeCacheSize, // number of keys to track frequency of (100K).
		MaxCost:     1 << 23,            // maximum cost of cache (3MB).
		// 100,000 roots will take up approximately 3 MB in memory.
		BufferItems: 64, // number of keys per Get buffer.
	})
	return &basicSSZ{
		hashCache: cache,
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
	if startOffset >= uint64(len(buf)) {
		return 0, fmt.Errorf("startOffset %d is greater than length of input %d", startOffset, len(buf))
	}

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

func (b *basicSSZ) Root(val reflect.Value, typ reflect.Type, fieldName string, maxCapacity uint64) ([32]byte, error) {
	var chunks [][]byte
	var err error
	var hashKey string
	newVal := reflect.New(val.Type()).Elem()
	newVal.Set(val)
	if val.Type().Kind() == reflect.Slice && val.IsNil() {
		newVal.Set(reflect.MakeSlice(val.Type(), typ.Len(), typ.Len()))
	}
	buf := make([]byte, DetermineSize(newVal))
	if _, err := b.Marshal(newVal, typ, buf, 0); err != nil {
		return [32]byte{}, err
	}
	hashKey = string(buf)
	res, ok := b.hashCache.Get(string(hashKey))
	if res != nil && ok {
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
	b.hashCache.Set(string(hashKey), root, 32)
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
	if val.Kind() == reflect.Array {
		for i := 0; i < val.Len(); i++ {
			buf[int(startOffset)+i] = uint8(val.Index(i).Uint())
		}
		return startOffset + uint64(val.Len()), nil
	}
	if val.IsNil() {
		item := make([]byte, typ.Len())
		copy(buf[startOffset:], item)
		return startOffset + uint64(typ.Len()), nil
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
