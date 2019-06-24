package ssz

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// field defines a custom wrapper around a struct field which
// include the respective sszUtils for that particular field type,
// giving easy access to its marshaler, unmarshaler, and tree hasher.
type field struct {
	index    int
	name     string
	typ      reflect.Type
	sszUtils *sszUtils
}

// truncateLast removes the last value of a struct, usually the signature,
// in order to hash only the data the signature field is intended to represent.
func truncateLast(typ reflect.Type) (fields []field, err error) {
	fields, err = structFields(typ)
	if err != nil {
		return nil, err
	}
	return fields[:len(fields)-1], nil
}

// structFields iterates over the raw fields of a struct, ignoring XXX protobuf fields,
// and determines the necessary ssz utils such as the marshaler, unmarshaler, and tree hasher
// for that particular struct field. Then, it returns a slice of field wrappers containing
// the necessary SSZ utils and field type information.
func structFields(typ reflect.Type) (fields []field, err error) {
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct kind input, received kind: %v", typ.Kind())
	}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if strings.Contains(f.Name, "XXX") {
			continue
		}
		// fieldType parses the struct's tags to check if there are any ssz tags
		// which specify a field should be treated as fixed-size by the marshaler.
		fType, err := fieldType(f)
		if err != nil {
			return nil, err
		}
		utils, err := cachedSSZUtilsNoAcquireLock(f.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to get ssz utils: %v", err)
		}
		name := f.Name
		fields = append(fields, field{index: i, name: name, sszUtils: utils, typ: fType})
	}
	return fields, nil
}

func sszTagSizes(tag string) ([]int, error) {
	sizeStartIndex := strings.IndexRune(tag, '=')
	items := strings.Split(tag[sizeStartIndex+1:], ",")
	sizes := make([]int, len(items))
	var err error
	for i := 0; i < len(items); i++ {
		if items[i] == "?" {
			// TODO: find a way to handle this nicely.
			sizes[i] = 0
			continue
		}
		sizes[i], err = strconv.Atoi(items[i])
		if err != nil {
			return nil, err
		}
	}
	return sizes, nil
}

func fieldType(field reflect.StructField) (reflect.Type, error) {
	item, exists := field.Tag.Lookup("ssz")
	if exists {
		sizes, err := sszTagSizes(item)
		if err != nil {
			return nil, err
		}
		if field.Type.Elem().Kind() == reflect.Slice {
			if len(sizes) > 1 {
				innerData := reflect.ArrayOf(sizes[1], field.Type.Elem().Elem())
				if sizes[0] == 0 {
					return reflect.SliceOf(innerData), nil
				}
				return reflect.ArrayOf(sizes[0], innerData), nil
			}
			innerData := reflect.ArrayOf(sizes[0], field.Type.Elem().Elem())
			return reflect.SliceOf(innerData), nil
		}
		return reflect.ArrayOf(sizes[0], field.Type.Elem()), nil
	}
	return field.Type, nil
}
