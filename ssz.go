package ssz

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/prysmaticlabs/go-ssz/types"
)

// Marshal a value and output the result into a byte slice.
// Given a struct with the following fields, one can marshal it as follows:
//  type exampleStruct struct {
//      Field1 uint8
//      Field2 []byte
//  }
//
//  ex := exampleStruct{
//      Field1: 10,
//      Field2: []byte{1, 2, 3, 4},
//  }
//  encoded, err := Marshal(ex)
//  if err != nil {
//      return fmt.Errorf("failed to marshal: %v", err)
//  }
//
// One can also specify the specific size of a struct's field by using
// ssz-specific field tags as follows:
//
//  type exampleStruct struct {
//      Field1 uint8
//      Field2 []byte `ssz:"size=32"`
//  }
//
// This will treat `Field2` as as [32]byte array when marshaling. For unbounded
// fields or multidimensional slices, ssz size tags can also be used as follows:
//
//  type exampleStruct struct {
//      Field1 uint8
//      Field2 [][]byte `ssz:"size=?,32"`
//  }
//
// This will treat `Field2` as type [][32]byte when marshaling a
// struct of that type.
func Marshal(val interface{}) ([]byte, error) {
	if val == nil {
		return nil, errors.New("untyped-value nil cannot be marshaled")
	}
	rval := reflect.ValueOf(val)

	// We pre-allocate a buffer-size depending on the value's calculated total byte size.
	buf := make([]byte, types.DetermineSize(rval))
	factory, err := types.SSZFactory(rval, rval.Type())
	if err != nil {
		return nil, err
	}
	if rval.Type().Kind() == reflect.Ptr {
		if _, err := factory.Marshal(rval.Elem(), rval.Elem().Type(), buf, 0 /* start offset */); err != nil {
			return nil, errors.Wrapf(err, "failed to marshal for type: %v", rval.Elem().Type())
		}
		return buf, nil
	}
	if _, err := factory.Marshal(rval, rval.Type(), buf, 0 /* start offset */); err != nil {
		return nil, errors.Wrapf(err, "failed to marshal for type: %v", rval.Type())
	}
	return buf, nil
}

// Unmarshal SSZ encoded data and output it into the object pointed by pointer val.
// Given a struct with the following fields, and some encoded bytes of type []byte,
// one can then unmarshal the bytes into a pointer of the struct as follows:
//  type exampleStruct1 struct {
//      Field1 uint8
//      Field2 []byte
//  }
//
//  var targetStruct exampleStruct1
//  if err := Unmarshal(encodedBytes, &targetStruct); err != nil {
//      return fmt.Errorf("failed to unmarshal: %v", err)
//  }
func Unmarshal(input []byte, val interface{}) error {
	if val == nil {
		return errors.New("cannot unmarshal into untyped, nil value")
	}
	if len(input) == 0 {
		return errors.New("no data to unmarshal from, input is an empty byte slice []byte{}")
	}
	rval := reflect.ValueOf(val)
	rtyp := rval.Type()
	// val must be a pointer, otherwise we refuse to unmarshal
	if rtyp.Kind() != reflect.Ptr {
		return errors.New("can only unmarshal into a pointer target")
	}
	if rval.IsNil() {
		return errors.New("cannot output to pointer of nil value")
	}
	factory, err := types.SSZFactory(rval.Elem(), rtyp.Elem())
	if err != nil {
		return err
	}
	if _, err := factory.Unmarshal(rval.Elem(), rval.Elem().Type(), input, 0); err != nil {
		return errors.Wrapf(err, "could not unmarshal input into type: %v", rval.Elem().Type())
	}
	return nil
}

// HashTreeRoot determines the root hash using SSZ's Merkleization.
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
//      return errors.Wrap(err, "failed to compute root")
//  }
func HashTreeRoot(val interface{}) ([32]byte, error) {
	if val == nil {
		return [32]byte{}, errors.New("untyped nil is not supported")
	}
	rval := reflect.ValueOf(val)
	factory, err := types.SSZFactory(rval, rval.Type())
	if err != nil {
		return [32]byte{}, errors.Wrapf(err, "could not generate tree hasher for type: %v", rval.Type())
	}
	return factory.Root(rval, rval.Type(), 0)
}

// HashTreeRootBitlist determines the root hash of a bitfield.Bitlist type using SSZ's Merkleization.
func HashTreeRootBitlist(bfield bitfield.Bitlist, maxCapacity uint64) ([32]byte, error) {
	return types.BitlistRoot(bfield, maxCapacity)
}

// HashTreeRootWithCapacity determines the root hash of a dynamic list
// using SSZ's Merkleization and applies a max capacity value when computing the root.
// If the input is not a slice, the function returns an error.
//
//  accountBalances := []uint64{1, 2, 3, 4}
//  root, err := HashTreeRootWithCapacity(accountBalances, 100) // Max 100 accounts.
//  if err != nil {
//      return errors.Wrap(err, "failed to compute root")
//  }
func HashTreeRootWithCapacity(val interface{}, maxCapacity uint64) ([32]byte, error) {
	if val == nil {
		return [32]byte{}, errors.New("untyped nil is not supported")
	}
	rval := reflect.ValueOf(val)
	if rval.Kind() != reflect.Slice {
		return [32]byte{}, fmt.Errorf("expected slice-kind input, received %v", rval.Kind())
	}
	factory, err := types.SSZFactory(rval, rval.Type())
	if err != nil {
		return [32]byte{}, errors.Wrapf(err, "could not generate tree hasher for type: %v", rval.Type())
	}
	return factory.Root(rval, rval.Type(), maxCapacity)
}

// SigningRoot truncates the last property of the struct passed in
// and returns its tree hash. This is done because the last property
// usually contains the signature that which this data is the root for.
func SigningRoot(val interface{}) ([32]byte, error) {
	if val == nil {
		return [32]byte{}, errors.New("value cannot be nil")
	}
	valObj := reflect.ValueOf(val)
	if valObj.Type().Kind() == reflect.Ptr {
		if valObj.IsNil() {
			return [32]byte{}, errors.New("nil pointer given")
		}
		elem := valObj.Elem()
		elemType := valObj.Elem().Type()
		totalFields := 0
		for i := 0; i < elemType.NumField(); i++ {
			// We skip protobuf related metadata fields.
			if strings.Contains(elemType.Field(i).Name, "XXX_") {
				continue
			}
			totalFields++
		}
		return types.StructFactory.FieldsHasher(elem, elemType, totalFields-1)
	}
	totalFields := 0
	for i := 0; i < valObj.Type().NumField(); i++ {
		// We skip protobuf related metadata fields.
		if strings.Contains(valObj.Type().Field(i).Name, "XXX_") {
			continue
		}
		totalFields++
	}
	return types.StructFactory.FieldsHasher(valObj, valObj.Type(), totalFields-1)
}
