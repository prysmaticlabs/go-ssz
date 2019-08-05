package ssz

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/prysmaticlabs/go-ssz/types"
)

var (
	useCache = false
)

// ToggleCache allows to programmatically enable/disable the hash tree root cache.
// TODO: Re-enable functionality.
func ToggleCache(enableTreeCache bool) {
	useCache = enableTreeCache
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
//      return fmt.Errorf("failed to compute root: %v", err)
//  }
func HashTreeRoot(val interface{}) ([32]byte, error) {
	if val == nil {
		return [32]byte{}, errors.New("untyped nil is not supported")
	}
	rval := reflect.ValueOf(val)
	factory, err := types.SSZFactory(rval, rval.Type())
	if err != nil {
		return [32]byte{}, fmt.Errorf("could generate tree hasher for type: %v: %v", rval.Type(), err)
	}
	return factory.Root(rval, rval.Type(), 0)
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
	factory, err := types.SSZFactory(rval, rval.Type())
	if err != nil {
		return [32]byte{}, fmt.Errorf("could generate tree hasher for type: %v: %v", rval.Type(), err)
	}
	return factory.Root(rval, rval.Type(), maxCapacity)
}
