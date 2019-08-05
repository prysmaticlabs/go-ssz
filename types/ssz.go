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
		default:
			return CompositeSliceFactory, nil
		}
	case kind == reflect.Array && isBasicTypeArray(typ.Elem(), typ.Elem().Kind()):
		return BasicArrayFactory, nil
	case kind == reflect.Array:
		return CompositeArrayFactory, nil
	case kind == reflect.Struct:
		return StructFactory, nil
	case kind == reflect.Ptr:
		if val.IsNil() {
			return &NilFactory{}, nil
		}
		return SSZFactory(val.Elem(), typ.Elem())
	default:
		return nil, nil
	}
}

type NilFactory struct{}

func (n *NilFactory) Marshal(val reflect.Value, typ reflect.Type, buf []byte, startOffset uint64) (uint64, error) {
	return startOffset, nil
}

func (n *NilFactory) Root(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	return [32]byte{}, nil
}
