# Simple Serialize (SSZ)

This package implements simple serialize algorithm specified in official Ethereum 2.0 [spec](https://github.com/ethereum/eth2.0-specs/blob/master/specs/simple-serialize.md).

[![Build Status](https://badge.buildkite.com/5945b9820092260cdc05fc0c736f50df313e15929dd0c864c4.svg?branch=master)](https://buildkite.com/prysmatic-labs/go-ssz)
[![Documentation](https://godoc.org/github.com/prysmatic-labs/go-ssz?status.svg)](http://godoc.org/github.com/prysmatic-labs/go-ssz)
[![codecov](https://codecov.io/gh/prysmaticlabs/go-ssz/branch/master/graph/badge.svg)](https://codecov.io/gh/prysmaticlabs/go-ssz)

## API

Our simple serialize API is designed to match the popular JSON marshal/unmarshal API from the Go standard library

```go
// Marshal val and output the result.
func Marshal(val interface{}) ([]byte, error)
```

```go
// Unmarshal data from input and output it into the object pointed by pointer val.
func Unmarshal(input []byte, val interface{}) error
```

### Tree Hashing

```go
// HashTreeRoot SSZ marshals a value and packs its serialized bytes into leaves of a Merkle trie -
// then, it determines the Merkle root of the trie.
func HashTreeRoot(val interface{}) ([32]byte, error)
````

## Usage

Say you have a struct like this

```go
type exampleStruct1 struct {
	Field1 uint8
	Field2 []byte
}
````

Now you can encode this object like this

```go
e1 := exampleStruct1{
    Field1: 10,
    Field2: []byte{1, 2, 3, 4},
}
encoded, err := Marshal(e1)
if err != nil {
    return fmt.Errorf("failed to marshal: %v", err)
}
```

One can also specify the specific size of a struct's field by using
ssz-specific field tags as follows:

```go
type exampleStruct struct {
    Field1 uint8
    Field2 []byte `ssz:"size=32"`
}
```

This will treat `Field2` as as [32]byte array when marshaling. For unbounded
fields or multidimensional slices, ssz size tags can also be used as follows:

```go
type exampleStruct struct {
    Field1 uint8
    Field2 [][]byte `ssz:"size=?,32"`
}
```

This will treat `Field2` as type [][32]byte when marshaling a
struct of that type.

Similarly, you can unmarshal encoded bytes into its original form:

```go
var e2 exampleStruct
if err = Unmarshal(encoded, &e2); err != nil {
    return fmt.Errorf("failed to unmarshal: %v", err)
}
reflect.DeepEqual(e1, e2) // Returns true as e2 now has the same content as e1.
```

To calculate tree-hash root of the object run:

```go
root, err := HashTreeRoot(e1)
if err != nil {
    return fmt.Errorf("failed to compute Merkle root: %v", err)
}
```

### Supported data types
- bool
- uint8
- uint16
- uint32
- uint64
- slice
- array
- struct
- pointer
