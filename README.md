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
e1 := &exampleStruct1{
    Field1: 10,
    Field2: []byte{1, 2, 3, 4},
}
wBuf := new(bytes.Buffer)
if err = e1.EncodeSSZ(wBuf); err != nil {
    return fmt.Errorf("failed to encode: %v", err)
}
encoding := wBuf.Bytes() // encoding becomes [0 0 0 9 10 0 0 0 4 1 2 3 4]
```

To calculate tree-hash of the object
```go
var hash [32]byte
if hash, err = e1.TreeHashSSZ(); err != nil {
    return fmt.Errorf("failed to hash: %v", err)
}
// hash stores the hashing result
```

Similarly, you can implement the `Decodable` interface for this struct

```go
func (e *exampleStruct1) DecodeSSZ(r io.Reader) error {
	return Decode(r, e)
}
```

Now you can decode to create new struct

```go
e2 := new(exampleStruct1)
rBuf := bytes.NewReader(encoding)
if err = e2.DecodeSSZ(rBuf); err != nil {
    return fmt.Errorf("failed to decode: %v", err)
}
// e2 now has the same content as e1
```

## Notes

### Supported data types
- uint8
- uint16
- uint32
- uint64
- slice
- array
- struct
- pointer
