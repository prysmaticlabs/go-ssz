package ssz

import (
	"reflect"
	"strings"
)

func isBasicType(kind reflect.Kind) bool {
	return kind == reflect.Bool ||
		kind == reflect.Uint8 ||
		kind == reflect.Uint16 ||
		kind == reflect.Uint32 ||
		kind == reflect.Uint64
}

func isBasicTypeArray(typ reflect.Type, kind reflect.Kind) bool {
	return kind == reflect.Array && isBasicType(typ.Elem().Kind())
}

func isBasicTypeSlice(typ reflect.Type, kind reflect.Kind) bool {
	return kind == reflect.Slice && isBasicType(typ.Elem().Kind())
}

func isVariableSizeType(val reflect.Value, kind reflect.Kind) bool {
	switch {
	case isBasicType(kind):
		return false
	case isBasicTypeArray(val.Type(), kind):
		return false
	case kind == reflect.Slice:
		return true
	case kind == reflect.Array:
		return isVariableSizeType(val.Elem(), val.Elem().Kind())
	case kind == reflect.Struct:
		for i := 0; i < val.Type().NumField(); i++ {
			if isVariableSizeType(val.Field(i), val.Field(i).Kind()) {
				return true
			}
		}
		return false
	}
	return false
}

func determineFixedSize(val reflect.Value, typ reflect.Type) uint64 {
	kind := typ.Kind()
	switch {
	case kind == reflect.Bool:
		return 1
	case kind == reflect.Uint8:
		return 1
	case kind == reflect.Uint16:
		return 2
	case kind == reflect.Uint32:
		return 4
	case kind == reflect.Uint64:
		return 8
	case kind == reflect.Array && typ.Elem().Kind() == reflect.Uint8:
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		return uint64(typ.Len())
	case kind == reflect.Array || kind == reflect.Slice:
		return determineFixedSize(val.Elem(), typ.Elem()) * uint64(typ.Len())
	case kind == reflect.Struct:
		totalSize := uint64(0)
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if strings.Contains(f.Name, "XXX") {
				continue
			}
			totalSize += determineFixedSize(val.Field(i), val.Field(i).Type())
		}
		return totalSize
	default:
		return 0
	}
	return 0
}

func determineVariableSize(val reflect.Value, typ reflect.Type) uint64 {
	kind := typ.Kind()
	switch {
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		return uint64(val.Len())
	case kind == reflect.Slice || kind == reflect.Array:
		totalSize := uint64(0)
		for i := 0; i < val.Len(); i++ {
			varSize := determineSize(val.Index(i))
			if isVariableSizeType(val.Index(i), val.Index(i).Kind()) {
				totalSize += varSize + uint64(BytesPerLengthOffset)
			} else {
				totalSize += varSize
			}
		}
		return totalSize
	case kind == reflect.Struct:
		totalSize := uint64(0)
		for i := 0; i < val.Type().NumField(); i++ {
			if isVariableSizeType(val.Field(i), val.Field(i).Kind()) {
				varSize := determineVariableSize(val.Field(i), val.Field(i).Type())
				totalSize += varSize + uint64(BytesPerLengthOffset)
			} else {
				varSize := determineVariableSize(val.Field(i), val.Field(i).Type())
				totalSize += varSize
			}
		}
		return totalSize
	default:
		return 0
	}
}

func determineSize(val reflect.Value) uint64 {
	if isVariableSizeType(val, val.Kind()) {
		return determineVariableSize(val, val.Type())
	}
	return determineFixedSize(val, val.Type())
}

