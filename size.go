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

func isVariableSizeType(val reflect.Value, typ reflect.Type) bool {
	kind := typ.Kind()
	switch {
	case isBasicType(kind):
		return false
	case isBasicTypeArray(typ, kind):
		return false
	case kind == reflect.Slice:
		return true
	case kind == reflect.Array:
		return isVariableSizeType(val, typ.Elem())
	case kind == reflect.Struct:
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if strings.Contains(f.Name, "XXX") {
				continue
			}
			fType, err := fieldType(f)
			if err != nil {
				return false
			}
			if isVariableSizeType(val.Field(i), fType) {
				return true
			}
		}
		return false
	case kind == reflect.Ptr:
		return isVariableSizeType(val.Elem(), val.Elem().Type())
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
		return uint64(typ.Len())
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		return uint64(val.Len())
	case kind == reflect.Array || kind == reflect.Slice:
		num := determineFixedSize(val, typ.Elem()) * uint64(typ.Len())
		return num
	case kind == reflect.Struct:
		totalSize := uint64(0)
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if strings.Contains(f.Name, "XXX") {
				continue
			}
			fType, err := fieldType(f)
			if err != nil {
				return 0
			}
			totalSize += determineFixedSize(val.Field(i), fType)
		}
		return totalSize
	default:
		return 0
	}
}

func determineVariableSize(val reflect.Value, typ reflect.Type) uint64 {
	kind := typ.Kind()
	switch {
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		return uint64(val.Len())
	case kind == reflect.Slice || kind == reflect.Array:
		totalSize := uint64(0)
		//fmt.Printf("Slice of type %v size %d\n", typ, val.Len())
		for i := 0; i < val.Len(); i++ {
			varSize := determineSize(val.Index(i))
			if isVariableSizeType(val.Index(i), typ.Elem()) {
				totalSize += varSize + uint64(BytesPerLengthOffset)
			} else {
				totalSize += varSize
			}
		}
		return totalSize
	case kind == reflect.Struct:
		totalSize := uint64(0)
		for i := 0; i < typ.NumField(); i++ {
			fType, err := fieldType(typ.Field(i))
			if err != nil {
				return 0
			}
			if isVariableSizeType(val.Field(i), fType) {
				//fmt.Println("--Variable field")
				//fmt.Println(typ.Field(i).Name)
				//fmt.Println(fType)
				varSize := determineVariableSize(val.Field(i), fType)
				//fmt.Println(varSize+uint64(BytesPerLengthOffset))
				totalSize += varSize + uint64(BytesPerLengthOffset)
				//fmt.Println(" ")
				//fmt.Println(" ")
			} else {
				//fmt.Println("--Fixed field")
				//fmt.Println(typ.Field(i).Name)
				//fmt.Println(fType)
				varSize := determineFixedSize(val.Field(i), fType)
				//fmt.Println(varSize)
				totalSize += varSize
				//fmt.Println(" ")
				//fmt.Println(" ")
			}
		}
		return totalSize
	case kind == reflect.Ptr:
		return determineVariableSize(val.Elem(), val.Elem().Type())
	default:
		return 0
	}
}

func determineSize(val reflect.Value) uint64 {
	if val.Kind() == reflect.Ptr {
		return determineSize(val.Elem())
	}
	if isVariableSizeType(val, val.Type()) {
		return determineVariableSize(val, val.Type())
	}
	return determineFixedSize(val, val.Type())
}
