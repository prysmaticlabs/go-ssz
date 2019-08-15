package types

import (
	"fmt"
	"reflect"
)

// StructFactory exports an implementation of a interface
// containing helpers for marshaling/unmarshaling, and determining
// the hash tree root of struct values.
var StructFactory = newStructSSZ()
var basicFactory = newBasicSSZ()
var basicArrayFactory = newBasicArraySSZ()
var compositeArrayFactory = newCompositeArraySSZ()
var basicSliceFactory = newBasicSliceSSZ()
var stringFactory = newStringSSZ()
var compositeSliceFactory = newCompositeSliceSSZ()

// SSZAble defines a type which can marshal/unmarshal and compute its
// hash tree root according to the Simple Serialize specification.
// See: https://github.com/ethereum/eth2.0-specs/blob/v0.8.2/specs/simple-serialize.md.
type SSZAble interface {
	Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error)
	Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error)
	Unmarshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error)
}

// SSZFactory recursively walks down a type and determines which SSZ-able
// core type it belongs to, and then returns and implementation of
// SSZ-able that contains marshal, unmarshal, and hash tree root related
// functions for use.
func SSZFactory(val reflect.Value, typ reflect.Type) (SSZAble, error) {
	kind := typ.Kind()
	switch {
	case isBasicType(kind) || isBasicTypeArray(typ, typ.Kind()):
		return basicFactory, nil
	case kind == reflect.String:
		return stringFactory, nil
	case kind == reflect.Slice:
		switch {
		case isBasicType(typ.Elem().Kind()):
			return basicSliceFactory, nil
		case !isVariableSizeType(typ.Elem()):
			return basicSliceFactory, nil
		default:
			return compositeSliceFactory, nil
		}
	case kind == reflect.Array:
		switch {
		case isBasicTypeArray(typ.Elem(), typ.Elem().Kind()):
			return basicArrayFactory, nil
		case !isVariableSizeType(typ.Elem()):
			return basicArrayFactory, nil
		default:
			return compositeArrayFactory, nil
		}
	case kind == reflect.Struct:
		return StructFactory, nil
	case kind == reflect.Ptr:
		return SSZFactory(val.Elem(), typ.Elem())
	default:
		return nil, fmt.Errorf("unsupported kind: %v", kind)
	}
}
