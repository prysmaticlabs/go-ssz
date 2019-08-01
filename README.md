[![Build Status](https://badge.buildkite.com/5945b9820092260cdc05fc0c736f50df313e15929dd0c864c4.svg?branch=master)](https://buildkite.com/prysmatic-labs/go-ssz)
[![codecov](https://codecov.io/gh/prysmaticlabs/go-ssz/branch/master/graph/badge.svg)](https://codecov.io/gh/prysmaticlabs/go-ssz)
[![Documentation](https://godoc.org/github.com/prysmatic-labs/go-ssz?status.svg)](http://godoc.org/github.com/prysmatic-labs/go-ssz)
[![Discord](https://user-images.githubusercontent.com/7288322/34471967-1df7808a-efbb-11e7-9088-ed0b04151291.png)](https://discord.gg/KSA7rPr)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/prysmaticlabs/geth-sharding?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

# Simple Serialize (SSZ)
Simple Serialize is the serialization algorithm standard for all data structures common across Ethereum 2.0 client implementations. It is outlined in the official [Ethereum 2.0 specification](https://github.com/ethereum/eth2.0-specs/blob/master/specs/simple-serialize.md). 

### Need assistance?
A more in-depth breakdown of SSZ is available in [this section](https://prysmaticlabs.gitbook.io/prysm/how-prysm-works/ethereum-2.0-serialization-simple-serialize-ssz) of our official documentation. If you still have questions, feel free to stop by either our [Discord](https://discord.gg/KSA7rPr) or [Gitter](https://gitter.im/prysmaticlabs/geth-sharding?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge) and a member of the team or our community will be happy to assist you.

## Functionality

### Marshal & Unmarshal
Our simple serialize API is designed to match the syntax of the popular JSON [marshal](https://golang.org/pkg/encoding/json/#Marshal) / [unmarshal](https://golang.org/pkg/encoding/json/#Unmarshal) API from the [Go](https://golang.org) standard library. Below are samples of each in use:

`marshal` example:
```go
// Marshal val and output the result.
func Marshal(val interface{}) ([]byte, error)
```
`unmarshal` example:
```go
// Unmarshal data from input and output it into the object pointed by pointer val.
func Unmarshal(input []byte, val interface{}) error
```

### Tree hashing
`HashTreeRoot` SSZ marshals a value and packs its serialized bytes into leaves of a [Merkle trie](https://github.com/ethereum/wiki/wiki/Patricia-Tree). It then determines the root of this trie.

`HashTreeRoot`example:
```go
func HashTreeRoot(val interface{}) ([32]byte, error)
````

## Usage examples
**Notice:** SSZ supports `bool`, `uint8`, `uint16`, `uint32`, `uint64`, `slice`, `array`, `struct` and `pointer` data types.

### Encoding an object (Marshal)

1. To begin, we create this basic struct:

```go
type exampleStruct1 struct {
    Field1 uint8
    Field2 []byte
}
````

2. Next, we can encode the defined object like so:

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

3. **(Optional)** It is also possible to specify the size of a struct's field by utilising ssz-specific field tags:

```go
type exampleStruct struct {
    Field1 uint8
    Field2 []byte `ssz:"size=32"`
}
```

This will treat `Field2` as as `[32]byte` array when marshaling. 

4. **(Optional)** For unbounded fields or multidimensional slices, SSZ size tags can also be used:

```go
type exampleStruct struct {
    Field1 uint8
    Field2 [][]byte `ssz:"size=?,32"`
}
```

This will treat `Field2` as type `[][32]byte` when marshaling a struct of that type.

### Decoding an object (Unmarshal)

1. Similarly, you can `unmarshal` encoded bytes into its original form:

```go
var e2 exampleStruct
if err = Unmarshal(encoded, &e2); err != nil {
    return fmt.Errorf("failed to unmarshal: %v", err)
}
reflect.DeepEqual(e1, e2) // Returns true as e2 now has the same content as e1.
```

### Calculating the tree-hash (HashTreeRoot)

1. To calculate tree-hash root of the object run:

```go
root, err := HashTreeRoot(e1)
if err != nil {
    return fmt.Errorf("failed to compute Merkle root: %v", err)
}
```

## Contributing
We have put all of our contribution guidelines into [CONTRIBUTING.md](https://github.com/prysmaticlabs/prysm/blob/master/CONTRIBUTING.md)! Check it out to get started.

## License

Brought to you by [Prysmatic Labs](https://prysmaticlabs.com) with a permissive [Apache License 2.0](license.md).
