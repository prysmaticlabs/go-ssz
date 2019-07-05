package ssz

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// UnboundedSSZFieldSizeMarker is the character used to specify a ssz field should have
// unbounded size, which is useful when describing slices of arrays such as [][32]byte.
// The ssz struct tag for such field type would be `ssz:"size=?,32"`. A question mark
// is chosen as the default value given its simplicity to represent unbounded size.
var UnboundedSSZFieldSizeMarker = "?"

// field defines a custom wrapper around a struct field which
// include the respective sszUtils for that particular field type,
// giving easy access to its marshaler, unmarshaler, and tree hasher.
type field struct {
	index       int
	name        string
	typ         reflect.Type
	sszUtils    *sszUtils
	capacity    uint64
	hasCapacity bool
	kind        string
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
		// determineFieldType parses the struct's tags to check if there are any ssz tags
		// which specify a field should be treated as fixed-size by the marshaler.
		fType, err := determineFieldType(f)
		if err != nil {
			return nil, err
		}
		fCapacity, hasCapacity := determineFieldCapacity(f)
		var fKind string
		if parsedKind, hasKind := determineFieldKind(f); hasKind {
			fKind = parsedKind
		}

		// We determine the SSZ utils for the field, including its respective
		// marshaler, unmarshaler, and hasher.
		utils, err := cachedSSZUtilsNoAcquireLock(fType)
		if err != nil {
			return nil, fmt.Errorf("failed to get ssz utils: %v", err)
		}
		name := f.Name
		fields = append(fields, field{
			index:       i,
			name:        name,
			sszUtils:    utils,
			typ:         fType,
			capacity:    fCapacity,
			hasCapacity: hasCapacity,
			kind:        fKind,
		})
	}
	return fields, nil
}

func determineFieldType(field reflect.StructField) (reflect.Type, error) {
	fieldSizeTags, exists, err := parseSSZFieldTags(field)
	if err != nil {
		return nil, fmt.Errorf("could not parse ssz struct field tags: %v", err)
	}
	if exists {
		// If the field does indeed specify ssz struct tags, we infer the field's type.
		return inferFieldTypeFromSizeTags(field, fieldSizeTags), nil
	}
	return field.Type, nil
}

func determineFieldCapacity(field reflect.StructField) (uint64, bool) {
	tag, exists := field.Tag.Lookup("ssz-max")
	if !exists {
		return 0, false
	}
	val, err := strconv.Atoi(tag)
	if err != nil {
		return 0, false
	}
	return uint64(val), true
}

func determineFieldKind(field reflect.StructField) (string, bool) {
	return field.Tag.Lookup("ssz-kind")
}

func parseSSZFieldTags(field reflect.StructField) ([]int, bool, error) {
	tag, exists := field.Tag.Lookup("ssz-size")
	if !exists {
		return nil, false, nil
	}
	items := strings.Split(tag, ",")
	sizes := make([]int, len(items))
	var err error
	for i := 0; i < len(items); i++ {
		// If a field is unbounded, we mark it with a size of 0.
		if items[i] == UnboundedSSZFieldSizeMarker {
			sizes[i] = 0
			continue
		}
		sizes[i], err = strconv.Atoi(items[i])
		if err != nil {
			return nil, false, err
		}
	}
	return sizes, true, nil
}

func inferFieldTypeFromSizeTags(field reflect.StructField, sizes []int) reflect.Type {
	innerElement := field.Type.Elem()
	for i := 1; i < len(sizes); i++ {
		innerElement = innerElement.Elem()
	}
	currentType := innerElement
	for i := len(sizes) - 1; i >= 0; i-- {
		if sizes[i] == 0 {
			currentType = reflect.SliceOf(currentType)
		} else {
			currentType = reflect.ArrayOf(sizes[i], currentType)
		}
	}
	return currentType
}

func growSliceFromSizeTags(val reflect.Value, sizes []int) reflect.Value {
	if len(sizes) == 0 {
		return val
	}
	finalValue := reflect.MakeSlice(val.Type(), sizes[0], sizes[0])
	for i := 0; i < sizes[0]; i++ {
		intermediate := growSliceFromSizeTags(finalValue.Index(i), sizes[1:])
		finalValue.Index(i).Set(intermediate)
	}
	return finalValue
}
