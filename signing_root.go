package ssz

import (
	"errors"
	"fmt"
	"reflect"
)

// SigningRoot truncates the last property of the struct passed in
// and returns its tree hash. This is done because the last property
// usually contains the signature that which this data is the root for.
func SigningRoot(val interface{}) ([32]byte, error) {
	valObj := reflect.ValueOf(val)
	kind := valObj.Kind()

	switch {
	case kind == reflect.Struct:
		return truncateAndHash(valObj)
	case kind == reflect.Ptr:
		if valObj.IsNil() {
			return [32]byte{}, errors.New("nil pointer given")
		}
		deRefVal := valObj.Elem()
		if deRefVal.Kind() != reflect.Struct {
			return [32]byte{}, errors.New("invalid type")
		}
		return truncateAndHash(deRefVal)
	default:
		return [32]byte{}, fmt.Errorf("given object is neither a struct or a pointer but is %v", kind)
	}
}

func truncateAndHash(val reflect.Value) ([32]byte, error) {
	truncated, err := truncateLast(val.Type())
	if err != nil {
		return [32]byte{}, err
	}
	hasher, err := makeFieldsHasher(truncated)
	if err != nil {
		return [32]byte{}, err
	}
	output, err := hasher(val, 0)
	if err != nil {
		return [32]byte{}, err
	}
	return output, nil
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
