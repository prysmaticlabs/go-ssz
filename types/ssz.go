package types

import "reflect"

var basicSSZ = NewBasicSSZ()
var compositeArraySSZ = NewCompositeArraySSZ()
var basicSliceSSZ = NewBasicSliceSSZ()
var compositeSliceSSZ = NewCompositeSliceSSZ()
var structSSZ = NewStructSSZ()

type SSZAble interface {
	Root(reflect.Value, reflect.Type, uint64) ([32]byte, error)
	Marshal(reflect.Value, reflect.Type, []byte, uint64) (uint64, error)
}

func SSZFactory(val reflect.Value, typ reflect.Type) (SSZAble, error) {
	kind := typ.Kind()
	switch {
	case isBasicType(kind) || isBasicTypeArray(typ, typ.Kind()):
		return basicSSZ, nil
	case kind == reflect.Slice:
		switch {
		case isBasicType(typ.Elem().Kind()):
			return basicSliceSSZ, nil
		default:
			return compositeSliceSSZ, nil
		}
	case kind == reflect.Array:
		return compositeArraySSZ, nil
	case kind == reflect.Struct:
		return structSSZ, nil
	default:
		return nil, nil
	}
}
