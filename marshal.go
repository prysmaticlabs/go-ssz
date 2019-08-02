package ssz

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"

	"github.com/prysmaticlabs/go-bitfield"
)

// Marshal a value and output the result into a byte slice.
// Given a struct with the following fields, one can marshal it as follows:
//  type exampleStruct struct {
//      Field1 uint8
//      Field2 []byte
//  }
//
//  ex := exampleStruct{
//      Field1: 10,
//      Field2: []byte{1, 2, 3, 4},
//  }
//  encoded, err := Marshal(ex)
//  if err != nil {
//      return fmt.Errorf("failed to marshal: %v", err)
//  }
//
// One can also specify the specific size of a struct's field by using
// ssz-specific field tags as follows:
//
//  type exampleStruct struct {
//      Field1 uint8
//      Field2 []byte `ssz:"size=32"`
//  }
//
// This will treat `Field2` as as [32]byte array when marshaling. For unbounded
// fields or multidimensional slices, ssz size tags can also be used as follows:
//
//  type exampleStruct struct {
//      Field1 uint8
//      Field2 [][]byte `ssz:"size=?,32"`
//  }
//
// This will treat `Field2` as type [][32]byte when marshaling a
// struct of that type.
func Marshal(val interface{}) ([]byte, error) {
	if val == nil {
		return nil, errors.New("untyped-value nil cannot be marshaled")
	}
	rval := reflect.ValueOf(val)

	// We pre-allocate a buffer-size depending on the value's calculated total byte size.
	buf := make([]byte, determineSize(rval))
	sszUtils, err := cachedSSZUtils(rval.Type())
	if err != nil {
		return nil, fmt.Errorf("could not initialize marshaler for type: %v", rval.Type())
	}
	if _, err = sszUtils.marshaler(rval, buf, 0 /* start offset */); err != nil {
		return nil, fmt.Errorf("failed to marshal for type: %v", rval.Type())
	}
	return buf, nil
}

func makeMarshaler(typ reflect.Type) (marshaler, error) {
	kind := typ.Kind()
	switch {
	case kind == reflect.Bool:
		return marshalBool, nil
	case kind == reflect.Uint8:
		return marshalUint8, nil
	case kind == reflect.Uint16:
		return marshalUint16, nil
	case kind == reflect.Uint32:
		return marshalUint32, nil
	case kind == reflect.Uint64:
		return marshalUint64, nil
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		return marshalByteSlice, nil
	case kind == reflect.Array && typ.Elem().Kind() == reflect.Uint8:
		return marshalByteArray, nil
	case kind == reflect.Slice && isBasicTypeArray(typ.Elem(), typ.Elem().Kind()):
		return makeBasicSliceMarshaler(typ)
	case kind == reflect.Slice && isBasicType(typ.Elem().Kind()):
		return makeBasicSliceMarshaler(typ)
	case kind == reflect.Slice && !isVariableSizeType(typ.Elem()):
		return makeBasicSliceMarshaler(typ)
	case kind == reflect.Slice || kind == reflect.Array:
		return makeCompositeSliceMarshaler(typ)
	case kind == reflect.Struct:
		return makeStructMarshaler(typ)
	case kind == reflect.Ptr:
		return makePtrMarshaler(typ)
	default:
		return nil, fmt.Errorf("type %v is not serializable", typ)
	}
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

func marshalByteSlice(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	if bits, ok := val.Interface().(bitfield.Bitlist); ok {
		copy(buf[startOffset:], bits[:])
		return startOffset + uint64(len(bits)), nil
	}
	v := val.Interface().([]byte)
	copy(buf[startOffset:], v)
	return startOffset + uint64(len(v)), nil
}

func marshalByteArray(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	switch v := val.Interface().(type) {
	case []uint8:
		copy(buf[startOffset:], v)
		return startOffset + uint64(len(v)), nil
	case bitfield.Bitvector4:
		copy(buf[startOffset:], v)
		return startOffset + uint64(len(v)), nil
	default:
		for i := 0; i < val.Len(); i++ {
			buf[int(startOffset)+i] = uint8(val.Index(i).Uint())
		}
		return startOffset + uint64(val.Len()), nil
	}
}

func makeBasicSliceMarshaler(typ reflect.Type) (marshaler, error) {
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, fmt.Errorf("failed to get ssz utils: %v", err)
	}

	marshaler := func(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
		index := startOffset
		var err error
		for i := 0; i < val.Len(); i++ {
			index, err = elemSSZUtils.marshaler(val.Index(i), buf, index)
			if err != nil {
				return 0, err
			}
		}
		return index, nil
	}
	return marshaler, nil
}

func makeCompositeSliceMarshaler(typ reflect.Type) (marshaler, error) {
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, fmt.Errorf("failed to get ssz utils: %v", err)
	}

	marshaler := func(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
		index := startOffset
		var err error
		if !isVariableSizeType(typ) {
			for i := 0; i < val.Len(); i++ {
				// If each element is not variable size, we simply encode sequentially and write
				// into the buffer at the last index we wrote at.
				index, err = elemSSZUtils.marshaler(val.Index(i), buf, index)
				if err != nil {
					return 0, err
				}
			}
		} else {
			fixedIndex := index
			currentOffsetIndex := startOffset + uint64(val.Len())*BytesPerLengthOffset
			nextOffsetIndex := currentOffsetIndex
			// If the elements are variable size, we need to include offset indices
			// in the serialized output list.
			for i := 0; i < val.Len(); i++ {
				nextOffsetIndex, err = elemSSZUtils.marshaler(val.Index(i), buf, currentOffsetIndex)
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
		}
		return index, nil
	}
	return marshaler, nil
}

func makeStructMarshaler(typ reflect.Type) (marshaler, error) {
	fields, err := structFields(typ)
	if err != nil {
		return nil, err
	}
	marshaler := func(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
		fixedIndex := startOffset
		fixedLength := uint64(0)
		// For every field, we add up the total length of the items depending if they
		// are variable or fixed-size fields.
		for _, f := range fields {
			if isVariableSizeType(f.typ) {
				fixedLength += BytesPerLengthOffset
			} else {
				fixedLength += determineFixedSize(val.Field(f.index), f.typ)
			}
		}
		currentOffsetIndex := startOffset + fixedLength
		nextOffsetIndex := currentOffsetIndex
		var err error
		for i, f := range fields {
			if !isVariableSizeType(f.typ) {
				fixedIndex, err = f.sszUtils.marshaler(val.Field(i), buf, fixedIndex)
				if err != nil {
					return 0, err
				}
			} else {
				nextOffsetIndex, err = f.sszUtils.marshaler(val.Field(f.index), buf, currentOffsetIndex)
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
		}
		return currentOffsetIndex, nil
	}
	return marshaler, nil
}

func makePtrMarshaler(typ reflect.Type) (marshaler, error) {
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, err
	}
	marshaler := func(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
		// Nil encodes to []byte{}.
		if val.IsNil() {
			return 0, nil
		}
		return elemSSZUtils.marshaler(val.Elem(), buf, startOffset)
	}

	return marshaler, nil
}
