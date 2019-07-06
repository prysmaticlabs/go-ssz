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
	sszUtils, err := cachedSSZUtils(rval.Type())
	if err != nil {
		return [32]byte{}, fmt.Errorf("could not get ssz utils for type: %v: %v", rval.Type(), err)
	}
	var output [32]byte
	if useCache {
		output, err = hashCache.lookup(rval, sszUtils.hasher)
	} else {
		output, err = sszUtils.hasher(rval, 0)
	}
	if err != nil {
		return [32]byte{}, fmt.Errorf("could not tree hash type: %v: %v", rval.Type(), err)
	}
	return output, nil
}

func makeHasher(typ reflect.Type) (hasher, error) {
	kind := typ.Kind()
	switch {
	case isBasicType(kind) || isBasicTypeArray(typ, kind):
		return makeBasicTypeHasher(typ)
	case kind == reflect.Slice && isBasicType(typ.Elem().Kind()):
		return makeBasicSliceHasher(typ)
	case kind == reflect.Slice && isBasicTypeArray(typ.Elem(), typ.Elem().Kind()):
		return makeBasicSliceHasher(typ)
	case kind == reflect.Slice && !isBasicType(typ.Elem().Kind()):
		return makeCompositeSliceHasher(typ)
	case kind == reflect.Array:
		return makeCompositeArrayHasher(typ)
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
	hasher := func(val reflect.Value, maxCapacity uint64) ([32]byte, error) {
		buf := make([]byte, determineSize(val))
		if _, err = utils.marshaler(val, buf, 0); err != nil {
			return [32]byte{}, err
		}
		chunks, err := pack([][]byte{buf})
		if err != nil {
			return [32]byte{}, err
		}
		result := merkleize(chunks, 0)
		return result, nil
	}
	return hasher, nil
}

func bitlistHasher(val reflect.Value, maxCapacity uint64) ([32]byte, error) {
	padding := (maxCapacity + 255) / 256
	buf := val.Interface().([]byte)
	bitfield := Bitlist(buf)
	chunks, err := pack([][]byte{bitfield.Bytes()})
	if err != nil {
		return [32]byte{}, err
	}
	length := make([]byte, 32)
	binary.PutUvarint(length, bitfield.Len())
	return mixInLength(merkleize(chunks, padding), length), nil
}

func makeCompositeArrayHasher(typ reflect.Type) (hasher, error) {
	utils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, err
	}
	hasher := func(val reflect.Value, maxCapacity uint64) ([32]byte, error) {
		roots := [][]byte{}
		for i := 0; i < val.Len(); i++ {
			var r [32]byte
			if useCache {
				r, err = hashCache.lookup(val.Index(i), utils.hasher)
			} else {
				r, err = utils.hasher(val.Index(i), 0)
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
		return merkleize(chunks, 0), nil
	}
	return hasher, nil
}

func makeBasicSliceHasher(typ reflect.Type) (hasher, error) {
	utils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, err
	}
	hasher := func(val reflect.Value, maxCapacity uint64) ([32]byte, error) {
		elemSize := uint64(0)
		if isBasicType(typ.Elem().Kind()) {
			elemSize = determineFixedSize(val, typ.Elem())
		} else {
			elemSize = 32
		}
		padding := (maxCapacity*elemSize + 31) / 32

		var leaves [][]byte
		for i := 0; i < val.Len(); i++ {
			if isBasicType(val.Index(i).Kind()) {
				innerBufSize := determineSize(val.Index(i))
				innerBuf := make([]byte, innerBufSize)
				if _, err = utils.marshaler(val.Index(i), innerBuf, 0); err != nil {
					return [32]byte{}, err
				}
				leaves = append(leaves, innerBuf)
			} else {
				r, err := utils.hasher(val.Index(i), 0)
				if err != nil {
					return [32]byte{}, err
				}
				leaves = append(leaves, r[:])
			}
		}
		chunks, err := pack(leaves)
		if err != nil {
			return [32]byte{}, err
		}
		buf := make([]byte, 32)
		binary.PutUvarint(buf, uint64(val.Len()))
		return mixInLength(merkleize(chunks, padding), buf), nil
	}
	return hasher, nil
}

func makeCompositeSliceHasher(typ reflect.Type) (hasher, error) {
	utils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, err
	}
	hasher := func(val reflect.Value, maxCapacity uint64) ([32]byte, error) {
		roots := [][]byte{}
		for i := 0; i < val.Len(); i++ {
			var r [32]byte
			if useCache {
				r, err = hashCache.lookup(val.Index(i), utils.hasher)
			} else {
				r, err = utils.hasher(val.Index(i), 0)
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
		buf := make([]byte, 32)
		binary.PutUvarint(buf, maxCapacity)
		return mixInLength(merkleize(chunks, 0), buf), nil
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
	hasher := func(val reflect.Value, maxCapacity uint64) ([32]byte, error) {
		fmt.Println("--Fields Hasher Running")
		roots := [][]byte{}
		for _, f := range fields {
			var r [32]byte
			var err error
			if useCache {
				r, err = hashCache.lookup(val.Field(f.index), f.sszUtils.hasher)
			} else {
				if f.kind == "bitlist" {
					r, err = bitlistHasher(val.Field(f.index), f.capacity)
				} else {
					if f.hasCapacity {
						r, err = f.sszUtils.hasher(val.Field(f.index), f.capacity)
					} else {
						r, err = f.sszUtils.hasher(val.Field(f.index), 0)
					}
				}
			}
			if err != nil {
				return [32]byte{}, fmt.Errorf("failed to hash field of struct: %v", err)
			}
			fmt.Printf("%v root %#x\n", f.name, r)
			roots = append(roots, r[:])
		}
		return merkleize(roots, 0), nil
	}
	return hasher, nil
}

func makePtrHasher(typ reflect.Type) (hasher, error) {
	elemSSZUtils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, err
	}
	hasher := func(val reflect.Value, maxCapacity uint64) ([32]byte, error) {
		if val.IsNil() {
			return [32]byte{}, nil
		}
		return elemSSZUtils.hasher(val.Elem(), maxCapacity)
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
