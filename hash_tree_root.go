package ssz

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
)

var useCache bool

// HashTreeRoot determines the root hash using SSZ's merkleization.
// Given a struct with the following fields, one can tree hash it as follows:
//  type exampleStruct struct {
//      Field1 uint8
//      Field2 []byte
//  }
//
//  ex := exampleStruct{
//      Field1: 10,
//      Field2: []byte{1, 2, 3, 4},
//  }
//  encoded, err := HashTreeRoot(ex)
//  if err != nil {
//      return fmt.Errorf("failed to marshal: %v", err)
//  }
func HashTreeRoot(val interface{}) ([32]byte, error) {
	if val == nil {
		return [32]byte{}, errors.New("untyped nil is not supported")
	}
	rval := reflect.ValueOf(val)
	sszUtils, err := cachedSSZUtilsNoAcquireLock(rval.Type())
	if err != nil {
		return [32]byte{}, fmt.Errorf("could not get ssz utils for type: %v: %v", rval.Type(), err)
	}
	var output [32]byte
	// if useCache {
	// 	output, err = hashCache.lookup(rval, sszUtils.hasher)
	// } else {
	output, err = sszUtils.hasher(rval)
	// }
	if err != nil {
		return [32]byte{}, fmt.Errorf("could not tree hash type: %v: %v", rval.Type(), err)
	}
	return output, nil
}

func makeHasher(typ reflect.Type) (hasher, error) {
	kind := typ.Kind()
	switch {
	// if the value is a basic object or an array of basic objects, we apply the basic
	// type hasher defined by merkleize(pack(value)).
	case isBasicType(kind) || isBasicTypeArray(typ, kind):
		return makeBasicTypeHasher(typ)
	// If the value is a slice of basic objects (dynamic length), we apply the basic slice
	// hasher defined by mix_in_length(merkleize(pack(value)), len(value)). Otherwise,
	// we apply mix_in_length(merkleize([hash_tree_root(element) for element in value]), len(value)).
	case kind == reflect.Slice:
		if isBasicTypeSlice(typ, kind) {
			return makeBasicSliceHasher(typ)
		}
		return makeCompositeSliceHasher(typ)
	// If the value is an array of composite objects, we apply the hasher
	// defined by merkleize([hash_tree_root(element) for element in value]).
	case kind == reflect.Array && !isBasicTypeArray(typ, kind):
		return makeCompositeArrayHasher(typ)
	// If the value is a container (a struct), we apply the struct hasher which is defined
	// by using the struct fields as merkleize([hash_tree_root(element) for element in value]).
	case kind == reflect.Struct:
		return makeStructHasher(typ)
	case kind == reflect.Ptr:
		return makePtrHasher(typ)
	default:
		return nil, fmt.Errorf("type %v is not hashable", typ)
	}
}

func makeBasicTypeHasher(typ reflect.Type) (hasher, error) {
	utils, err := cachedSSZUtilsNoAcquireLock(typ)
	if err != nil {
		return nil, err
	}
	hasher := func(val reflect.Value) ([32]byte, error) {
		buf := make([]byte, determineSize(val))
		if _, err = utils.marshaler(val, buf, 0); err != nil {
			return [32]byte{}, err
		}
		chunks, err := pack([][]byte{buf})
		if err != nil {
			return [32]byte{}, err
		}
		return merkleize(chunks), nil
	}
	return hasher, nil
}

func makeBasicSliceHasher(typ reflect.Type) (hasher, error) {
	utils, err := cachedSSZUtilsNoAcquireLock(typ)
	if err != nil {
		return nil, fmt.Errorf("failed to get ssz utils: %v", err)
	}
	hasher := func(val reflect.Value) ([32]byte, error) {
		buf := make([]byte, determineSize(val))
		if _, err = utils.marshaler(val, buf, 0); err != nil {
			return [32]byte{}, err
		}
		serializedValues := [][]byte{buf}
		chunks, err := pack(serializedValues)
		if err != nil {
			return [32]byte{}, err
		}
		// We marshal the length into little-endian, 256-bit byte slice.
		b := make([]byte, 32)
		binary.LittleEndian.PutUint64(b, uint64(val.Len()))
		return mixInLength(merkleize(chunks), b), nil
	}
	return hasher, nil
}

func makeCompositeSliceHasher(typ reflect.Type) (hasher, error) {
	utils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, err
	}
	hasher := func(val reflect.Value) ([32]byte, error) {
		roots := [][]byte{}
		for i := 0; i < val.Len(); i++ {
			var r [32]byte
			if useCache {
				r, err = hashCache.lookup(val.Index(i), utils.hasher)
			} else {
				r, err = utils.hasher(val.Index(i))
			}
			if err != nil {
				return [32]byte{}, err
			}
			roots = append(roots, r[:])
		}
		chunks, err := pack(roots)
		if err != nil {
			return [32]byte{}, err
		}
		b := make([]byte, 32)
		binary.LittleEndian.PutUint64(b, uint64(val.Len()))
		return mixInLength(merkleize(chunks), b), nil
	}
	return hasher, nil
}

func makeCompositeArrayHasher(typ reflect.Type) (hasher, error) {
	utils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, err
	}
	hasher := func(val reflect.Value) ([32]byte, error) {
		roots := [][]byte{}
		for i := 0; i < val.Len(); i++ {
			var r [32]byte
			if useCache {
				r, err = hashCache.lookup(val.Index(i), utils.hasher)
			} else {
				r, err = utils.hasher(val.Index(i))
			}
			if err != nil {
				return [32]byte{}, err
			}
			roots = append(roots, r[:])
		}
		chunks, err := pack(roots)
		if err != nil {
			return [32]byte{}, err
		}
		return merkleize(chunks), nil
	}
	return hasher, nil
}

func makeStructHasher(typ reflect.Type) (hasher, error) {
	fields, err := structFields(typ)
	if err != nil {
		return nil, err
	}
	return makeFieldsHasher(fields)
}

func makeFieldsHasher(fields []field) (hasher, error) {
	hasher := func(val reflect.Value) ([32]byte, error) {
		roots := [][]byte{}
		for _, f := range fields {
			var r [32]byte
			var err error
			if useCache {
				r, err = hashCache.lookup(val.Field(f.index), f.sszUtils.hasher)
			} else {
				r, err = f.sszUtils.hasher(val.Field(f.index))
			}
			if err != nil {
				return [32]byte{}, fmt.Errorf("failed to hash field of struct: %v", err)
			}
			roots = append(roots, r[:])
		}
		return merkleize(roots), nil
	}
	return hasher, nil
}

func makePtrHasher(typ reflect.Type) (hasher, error) {
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, err
	}
	hasher := func(val reflect.Value) ([32]byte, error) {
		if val.IsNil() {
			return [32]byte{}, nil
		}
		return elemSSZUtils.hasher(val.Elem())
	}
	return hasher, nil
}

func getEncoding(val reflect.Value) ([]byte, error) {
	utils, err := cachedSSZUtilsNoAcquireLock(val.Type())
	if err != nil {
		return nil, err
	}
	buf := make([]byte, determineSize(val))
	if _, err = utils.marshaler(val, buf, 0); err != nil {
		return nil, err
	}
	return buf, nil
}

// HashedEncoding returns the hash of the encoded object.
func HashedEncoding(val interface{}) ([32]byte, error) {
	rval := reflect.ValueOf(val)
	encoding, err := getEncoding(rval)
	if err != nil {
		return [32]byte{}, err
	}
	return hash(encoding), nil
}
