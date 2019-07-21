package ssz

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"

	"github.com/prysmaticlabs/go-bitfield"
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
//  root, err := HashTreeRoot(ex)
//  if err != nil {
//      return fmt.Errorf("failed to compute root: %v", err)
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

// HashTreeRootWithCapacity determines the root hash of a dynamic list
// using SSZ's merkleization and applies a max capacity value when computing the root.
// If the input is not a slice, the function returns an error.
//
//  accountBalances := []uint64{1, 2, 3, 4}
//  root, err := HashTreeRootWithCapacity(accountBalances, 100) // Max 100 accounts.
//  if err != nil {
//      return fmt.Errorf("failed to compute root: %v", err)
//  }
func HashTreeRootWithCapacity(val interface{}, maxCapacity uint64) ([32]byte, error) {
	if val == nil {
		return [32]byte{}, errors.New("untyped nil is not supported")
	}
	rval := reflect.ValueOf(val)
	if rval.Kind() != reflect.Slice {
		return [32]byte{}, fmt.Errorf("expected slice-kind input, received %v", rval.Kind())
	}
	sszUtils, err := cachedSSZUtils(rval.Type())
	if err != nil {
		return [32]byte{}, fmt.Errorf("could not get ssz utils for type: %v: %v", rval.Type(), err)
	}
	var output [32]byte
	if useCache {
		output, err = hashCache.lookup(rval, sszUtils.hasher)
	} else {
		output, err = sszUtils.hasher(rval, maxCapacity)
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
	case kind == reflect.Array && isBasicTypeArray(typ.Elem(), typ.Elem().Kind()):
		return makeBasicArrayHasher(typ)
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
		return bitwiseMerkleize(chunks, 1, false /* has limit */)
	}
	return hasher, nil
}

func bitlistHasher(val reflect.Value, maxCapacity uint64) ([32]byte, error) {
	limit := (maxCapacity + 255) / 256
	if val.IsNil() {
		length := make([]byte, 32)
		merkleRoot, err := bitwiseMerkleize([][]byte{}, limit, true /* has limit */)
		if err != nil {
			return [32]byte{}, err
		}
		return mixInLength(merkleRoot, length), nil
	}
	bfield := val.Interface().(bitfield.Bitlist)
	chunks, err := pack([][]byte{bfield.Bytes()})
	if err != nil {
		return [32]byte{}, err
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, bfield.Len())
	output := make([]byte, 32)
	copy(output, buf.Bytes())
	merkleRoot, err := bitwiseMerkleize(chunks, limit, true /* has limit */)
	if err != nil {
		return [32]byte{}, err
	}
	return mixInLength(merkleRoot, output), nil
}

func makeBasicArrayHasher(typ reflect.Type) (hasher, error) {
	utils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, err
	}
	hasher := func(val reflect.Value, maxCapacity uint64) ([32]byte, error) {
		var leaves [][]byte
		for i := 0; i < val.Len(); i++ {
			r, err := utils.hasher(val.Index(i), 0)
			if err != nil {
				return [32]byte{}, err
			}
			leaves = append(leaves, r[:])
		}
		chunks, err := pack(leaves)
		if err != nil {
			return [32]byte{}, err
		}
		if val.Len() == 0 {
			chunks = [][]byte{}
		}
		return bitwiseMerkleize(chunks, 1, false /* has limit */)
	}
	return hasher, nil
}

func makeCompositeArrayHasher(typ reflect.Type) (hasher, error) {
	utils, err := cachedSSZUtilsNoAcquireLock(typ.Elem())
	if err != nil {
		return nil, err
	}
	hasher := func(val reflect.Value, maxCapacity uint64) ([32]byte, error) {
		roots := [][]byte{}
		elemSize := uint64(0)
		if isBasicType(typ.Elem().Kind()) {
			elemSize = determineFixedSize(val, typ.Elem())
		} else {
			elemSize = 32
		}
		limit := (uint64(val.Len())*elemSize + 31) / 32
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
		if val.Len() == 0 {
			chunks = [][]byte{}
		}
		return bitwiseMerkleize(chunks, limit, true /* has limit */)
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
		limit := (maxCapacity*elemSize + 31) / 32
		if limit == 0 {
			limit = 1
		}

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
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint64(val.Len()))
		output := make([]byte, 32)
		copy(output, buf.Bytes())
		merkleRoot, err := bitwiseMerkleize(chunks, limit, true /* has limit */)
		if err != nil {
			return [32]byte{}, err
		}
		return mixInLength(merkleRoot, output), nil
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
		output := make([]byte, 32)
		if val.Len() == 0 && maxCapacity == 0 {
			merkleRoot, err := bitwiseMerkleize([][]byte{}, 0, true /* has limit */)
			if err != nil {
				return [32]byte{}, err
			}
			itemMerkleize := mixInLength(merkleRoot, output)
			return itemMerkleize, nil
		}
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
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint64(val.Len()))
		copy(output, buf.Bytes())
		objLen := maxCapacity
		if maxCapacity == 0 {
			objLen = uint64(val.Len())
		}
		merkleRoot, err := bitwiseMerkleize(chunks, objLen, true /* has limit */)
		if err != nil {
			return [32]byte{}, err
		}
		return mixInLength(merkleRoot, output), nil
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
		roots := [][]byte{}
		for _, f := range fields {
			var r [32]byte
			var err error
			if useCache {
				r, err = hashCache.lookup(val.Field(f.index), f.sszUtils.hasher)
			} else {
				if _, ok := val.Field(f.index).Interface().(bitfield.Bitlist); ok {
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
				return [32]byte{}, fmt.Errorf("failed to hash field %s of struct: %v", f.name, err)
			}
			roots = append(roots, r[:])
		}
		return bitwiseMerkleize(roots, uint64(len(fields)), true /* has limit */)
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
