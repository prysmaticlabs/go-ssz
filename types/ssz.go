package types

import (
	"reflect"
)

var StructFactory = NewStructSSZ()
var BasicFactory = NewBasicSSZ()
var BasicArrayFactory = NewBasicArraySSZ()
var CompositeArrayFactory = NewCompositeArraySSZ()
var BasicSliceFactory = NewBasicSliceSSZ()
var CompositeSliceFactory = NewCompositeSliceSSZ()

type SSZAble interface {
	Root(reflect.Value, reflect.Type, uint64) ([32]byte, error)
	Marshal(reflect.Value, reflect.Type, []byte, uint64) (uint64, error)
	Unmarshal(reflect.Value, reflect.Type, []byte, uint64) (uint64, error)
}

func SSZFactory(val reflect.Value, typ reflect.Type) (SSZAble, error) {
	kind := typ.Kind()
	switch {
	case isBasicType(kind) || isBasicTypeArray(typ, typ.Kind()):
		return BasicFactory, nil
	case kind == reflect.Slice:
		switch {
		case isBasicType(typ.Elem().Kind()):
			return BasicSliceFactory, nil
		case !isVariableSizeType(typ.Elem()):
			return BasicSliceFactory, nil
		default:
			return CompositeSliceFactory, nil
		}
	case kind == reflect.Array:
		switch {
		case isBasicTypeArray(typ.Elem(), typ.Elem().Kind()):
			return BasicArrayFactory, nil
		case !isVariableSizeType(typ.Elem()):
			return BasicArrayFactory, nil
		default:
			return CompositeArrayFactory, nil
		}
	case kind == reflect.Struct:
		return StructFactory, nil
	case kind == reflect.Ptr:
		return SSZFactory(val.Elem(), typ.Elem())
	default:
		return nil, nil
	}
}
