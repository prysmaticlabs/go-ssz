package types

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"time"

	"github.com/karlseguin/ccache"
)

type BasicSSZ struct {
	hashCache *ccache.Cache
}

func NewBasicSSZ() *BasicSSZ {
	return &BasicSSZ{
		hashCache: ccache.New(ccache.Configure().MaxSize(100000)),
	}
}

func (b *BasicSSZ) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	kind := typ.Kind()
	switch {
	case kind == reflect.Bool:
		return marshalBool(val, buf, startOffset)
	case kind == reflect.Uint8:
		return marshalUint8(val, buf, startOffset)
	case kind == reflect.Uint16:
		return marshalUint16(val, buf, startOffset)
	case kind == reflect.Uint32:
		return marshalUint32(val, buf, startOffset)
	case kind == reflect.Uint64:
		return marshalUint64(val, buf, startOffset)
	case kind == reflect.Array && isBasicType(typ.Elem().Kind()):
		return b.marshalBasicSlice(val, typ, buf, startOffset)
	default:
		return 0, fmt.Errorf("type %v is not serializable", val.Type())
	}
}

func (b *BasicSSZ) Root(val reflect.Value, typ reflect.Type) ([32]byte, error) {
	var chunks [][]byte
	var err error
	var hashKey string
	buf := make([]byte, determineSize(val))
	if _, err := b.Marshal(val, typ, buf, 0); err != nil {
		return [32]byte{}, err
	}
	hashKey = string(buf)
	res := b.hashCache.Get(hashKey)
	if res != nil {
		return res.Value().([32]byte), nil
	}

	chunks, err = pack([][]byte{buf})
	if err != nil {
		return [32]byte{}, err
	}
	root, err := bitwiseMerkleize(chunks, 1, false /* has limit */)
	if err != nil {
		return [32]byte{}, err
	}
	b.hashCache.Set(hashKey, root, time.Hour)
	return root, nil
}

func (b *BasicSSZ) marshalBasicSlice(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
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

func marshalBool(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	if val.Interface().(bool) {
		buf[startOffset] = uint8(1)
	} else {
		buf[startOffset] = uint8(0)
	}
	return startOffset + 1, nil
}

func marshalUint8(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	buf[startOffset] = val.Interface().(uint8)
	return startOffset + 1, nil
}

func marshalUint16(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	binary.LittleEndian.PutUint16(buf[startOffset:], val.Interface().(uint16))
	return startOffset + 2, nil
}

func marshalUint32(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	binary.LittleEndian.PutUint32(buf[startOffset:], val.Interface().(uint32))
	return startOffset + 4, nil
}

func marshalUint64(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	binary.LittleEndian.PutUint64(buf[startOffset:], val.Interface().(uint64))
	return startOffset + 8, nil
}
