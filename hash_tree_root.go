package ssz

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"

	"github.com/prysmaticlabs/go-bitfield"
	"github.com/prysmaticlabs/go-ssz/types"
)

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
func NewHashTreeRoot(val interface{}) ([32]byte, error) {
	if val == nil {
		return [32]byte{}, errors.New("untyped nil is not supported")
	}
	rval := reflect.ValueOf(val)
	var r [32]byte
	var err error
	r, err = recursiveHasher(rval, rval.Type(), 0)
	if err != nil {
		return [32]byte{}, fmt.Errorf("could not tree hash type: %v: %v", rval.Type(), err)
	}
	return r, nil
}

func recursiveHasher(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	kind := typ.Kind()
	switch {
	case isBasicType(kind) || isBasicTypeArray(typ, kind):
		return basicTypeHasher(val, typ)
	case kind == reflect.Slice && isBasicType(typ.Elem().Kind()):
		return basicSliceHasher(val, typ, maxCapacity)
	case kind == reflect.Slice && isBasicTypeArray(typ.Elem(), typ.Elem().Kind()):
		return basicSliceHasher(val, typ, maxCapacity)
	case kind == reflect.Array && isBasicTypeArray(typ.Elem(), typ.Elem().Kind()):
		return basicArrayHasher(val, typ)
	case kind == reflect.Slice && !isBasicType(typ.Elem().Kind()):
		return compositeSliceHasher(val, typ, maxCapacity)
	case kind == reflect.Array:
		return compositeArrayHasher(val, typ)
	case kind == reflect.Struct:
		return structHasher(val, typ)
	case kind == reflect.Ptr:
		return ptrHasher(val, typ, maxCapacity)
	default:
		return [32]byte{}, fmt.Errorf("type %v is not hashable", typ)
	}
}

func basicTypeHasher(val reflect.Value, typ reflect.Type) ([32]byte, error) {
	var chunks [][]byte
	var err error
	if b, ok := val.Interface().([]byte); ok {
		chunks, err = pack([][]byte{b})
		if err != nil {
			return [32]byte{}, err
		}
	} else {
		buf := make([]byte, determineSize(val))
		if _, err := types.BasicFactory.Marshal(val, typ, buf, 0); err != nil {
			return [32]byte{}, err
		}
		chunks, err = pack([][]byte{buf})
		if err != nil {
			return [32]byte{}, err
		}
	}
	return bitwiseMerkleize(chunks, 1, false /* has limit */)
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
	if err := binary.Write(buf, binary.LittleEndian, bfield.Len()); err != nil {
		return [32]byte{}, err
	}
	output := make([]byte, 32)
	copy(output, buf.Bytes())
	merkleRoot, err := bitwiseMerkleize(chunks, limit, true /* has limit */)
	if err != nil {
		return [32]byte{}, err
	}
	return mixInLength(merkleRoot, output), nil
}

func basicArrayHasher(val reflect.Value, typ reflect.Type) ([32]byte, error) {
	var leaves [][]byte
	var err error
	for i := 0; i < val.Len(); i++ {
		r, err := recursiveHasher(val.Index(i), typ.Elem(), 0)
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

func compositeArrayHasher(val reflect.Value, typ reflect.Type) ([32]byte, error) {
	roots := [][]byte{}
	elemSize := uint64(0)
	if isBasicType(typ.Elem().Kind()) {
		elemSize = determineFixedSize(val, typ.Elem())
	} else {
		elemSize = 32
	}
	limit := (uint64(val.Len())*elemSize + 31) / 32
	var err error
	for i := 0; i < val.Len(); i++ {
		r, err := recursiveHasher(val.Index(i), typ.Elem(), 0)
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

func basicSliceHasher(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
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
	var err error
	for i := 0; i < val.Len(); i++ {
		if isBasicType(val.Index(i).Kind()) {
			innerBuf := make([]byte, elemSize)
			if _, err := types.BasicFactory.Marshal(val.Index(i), typ.Elem(), innerBuf, 0); err != nil {
				return [32]byte{}, fmt.Errorf("fails here: %v", err)
			}
			leaves = append(leaves, innerBuf)
		} else {
			r, err := recursiveHasher(val.Index(i), typ.Elem(), 0)
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
	if err := binary.Write(buf, binary.LittleEndian, uint64(val.Len())); err != nil {
		return [32]byte{}, err
	}
	output := make([]byte, 32)
	copy(output, buf.Bytes())
	merkleRoot, err := bitwiseMerkleize(chunks, limit, true /* has limit */)
	if err != nil {
		return [32]byte{}, err
	}
	return mixInLength(merkleRoot, output), nil
}

func compositeSliceHasher(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
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
	var err error
	for i := 0; i < val.Len(); i++ {
		r, err := recursiveHasher(val.Index(i), typ.Elem(), 0)
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
	if err := binary.Write(buf, binary.LittleEndian, uint64(val.Len())); err != nil {
		return [32]byte{}, err
	}

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

func structHasher(val reflect.Value, typ reflect.Type) ([32]byte, error) {
	fields, err := structFields(typ)
	if err != nil {
		return [32]byte{}, err
	}
	return fieldsHasher(val, fields)
}

func fieldsHasher(val reflect.Value, fields []field) ([32]byte, error) {
	roots := make([][]byte, len(fields))
	for i, f := range fields {
		var r [32]byte
		var err error
		if _, ok := val.Field(f.index).Interface().(bitfield.Bitlist); ok {
			r, err = bitlistHasher(val.Field(f.index), f.capacity)
			roots[i] = r[:]
			continue
		}
		r, err = recursiveHasher(val.Field(f.index), f.typ, f.capacity)
		if err != nil {
			return [32]byte{}, fmt.Errorf("failed to hash field %s of struct: %v", f.name, err)
		}
		fmt.Printf("%s has root %#x\n", f.name, r)
		roots[i] = r[:]
	}
	return bitwiseMerkleize(roots, uint64(len(fields)), true /* has limit */)
}

func ptrHasher(val reflect.Value, typ reflect.Type, maxCapacity uint64) ([32]byte, error) {
	if val.IsNil() {
		return [32]byte{}, nil
	}
	return recursiveHasher(val.Elem(), typ.Elem(), maxCapacity)
}