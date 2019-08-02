package types

import "reflect"

var basicSSZ = NewBasicSSZ()
var compositeSliceSSZ = NewCompositeSliceSSZ()
var basicSliceSSZ = NewBasicSliceSSZ()

type SSZAble interface {
	Root(reflect.Value, reflect.Type) ([32]byte, error)
}

func SSZFactory(val reflect.Value, typ reflect.Type) (SSZAble, error) {
	kind := typ.Kind()
	switch {
	case isBasicType(kind) || isBasicTypeArray(typ, typ.Kind()):
		return basicSSZ, nil
	case kind == reflect.Slice:
		switch {
		case isVariableSizeType(val.Type().Elem()):
			return compositeSliceSSZ, nil
		default:
			return basicSliceSSZ, nil
		}
	default:
		return nil, nil
	}
}
