package ssz

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"sync"
	"testing"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-bitfield"
)

type fork struct {
	PreviousVersion [4]byte
	CurrentVersion  [4]byte
	Epoch           uint64
}

type truncateSignatureCase struct {
	Slot              uint64
	PreviousBlockRoot []byte
	Signature         []byte
}

type simpleNonProtoMessage struct {
	Foo []byte
	Bar uint64
}

// This test verifies if a nil pseudo-array is treated the same as an instantiated,
// zero-valued array when running hash tree root computations.
func TestEmptyArrayInstantiation(t *testing.T) {
	type data struct {
		DepositRoot  []byte `ssz-size:"32"`
		DepositCount uint64
		BlockHash    []byte `ssz-size:"32"`
	}
	type example struct {
		Randao   []byte `ssz-size:"96"`
		Data     *data
		Graffiti []byte `ssz-size:"32"`
	}
	empty := &example{
		Randao: make([]byte, 96),
		Data: &data{
			DepositRoot:  make([]byte, 32),
			DepositCount: 0,
			BlockHash:    make([]byte, 32),
		},
	}
	withInstantiatedArray := &example{
		Randao: make([]byte, 96),
		Data: &data{
			DepositRoot:  make([]byte, 32),
			DepositCount: 0,
			BlockHash:    make([]byte, 32),
		},
		Graffiti: make([]byte, 32),
	}
	r1, err := HashTreeRoot(empty)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := HashTreeRoot(withInstantiatedArray)
	if err != nil {
		t.Fatal(err)
	}
	if r1 != r2 {
		t.Errorf("Wanted nil_array_field = %#x, instiantiated_empty_array_field = %#x", r1, r2)
	}
}

func TestMarshalNilArray(t *testing.T) {
	type ex struct {
		Slot         uint64
		Graffiti     []byte `ssz-size:"32"`
		DepositIndex uint64
	}
	b1 := &ex{
		Slot:         5,
		Graffiti:     nil,
		DepositIndex: 64,
	}
	b2 := &ex{
		Slot:         5,
		Graffiti:     make([]byte, 32),
		DepositIndex: 64,
	}
	enc1, err := Marshal(b1)
	if err != nil {
		t.Fatal(err)
	}
	enc2, err := Marshal(b2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(enc1, enc2) {
		t.Errorf("First item %v != second item %v", enc1, enc2)
	}
}

func TestPartialDataMarshalUnmarshal(t *testing.T) {
	type block struct {
		Slot      uint64
		Transfers []*simpleProtoMessage
	}
	b := &block{
		Slot: 5,
	}
	enc, err := Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	dec := &block{}
	if err := Unmarshal(enc, dec); err != nil {
		t.Fatal(err)
	}
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		output []byte
		err    error
	}{
		{
			name: "Nil",
			err:  errors.New("untyped-value nil cannot be marshaled"),
		},
		{
			name:  "Unsupported",
			input: complex(1, 1),
			err:   errors.New("unsupported kind: complex128"),
		},
		{
			name:  "UnsupportedPointer",
			input: &[]complex128{complex(1, 1), complex(1, 1)},
			err:   errors.New("failed to marshal for type: []complex128: unsupported kind: complex128"),
		},
		{
			name:  "UnsupportedStructElement",
			input: struct{ Foo complex128 }{complex(1, 1)},
			err:   errors.New("failed to marshal for type: struct { Foo complex128 }: unsupported kind: complex128"),
		},
		{
			name:   "Simple",
			input:  struct{ Foo uint32 }{12345},
			output: []byte{0x39, 0x30, 0x00, 0x00},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := Marshal(test.input)
			if test.err == nil {
				if err != nil {
					t.Fatalf("unexpected error %v", err)
				}
				if bytes.Compare(test.output, output) != 0 {
					t.Errorf("incorrect output: expected %v; received %v", test.output, output)
				}
			} else {
				if err == nil {
					t.Fatalf("missing expected error %v", test.err)
				}
				if test.err.Error() != err.Error() {
					t.Errorf("incorrect error: expected %v; received %v", test.err, err)
				}
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		output interface{}
		err    error
	}{
		{
			name: "Nil",
			err:  errors.New("cannot unmarshal into untyped, nil value"),
		},
		{
			name:   "NotPointer",
			input:  []byte{0x00, 0x00, 0x00, 0x00},
			output: "",
			err:    errors.New("can only unmarshal into a pointer target"),
		},
		{
			name:   "OutputNotSupported",
			input:  []byte{0x00, 0x00, 0x00, 0x00},
			output: &struct{ Foo complex128 }{complex(1, 1)},
			err:    errors.New("could not unmarshal input into type: struct { Foo complex128 }: unsupported kind: complex128"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Unmarshal(test.input, test.output)
			if test.err == nil {
				if err != nil {
					t.Errorf("unexpected error %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("missing expected error %v", test.err)
				}
				if test.err.Error() != err.Error() {
					t.Errorf("unexpected error value %v (expected %v)", err, test.err)
				}
			}
		})
	}
}

func TestHashTreeRoot(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		output [32]byte
		err    error
	}{
		{
			name: "Nil",
			err:  errors.New("untyped nil is not supported"),
		},
		{
			name:  "UnsupportedKind",
			input: complex(1, 1),
			err:   errors.New("could not generate tree hasher for type: complex128: unsupported kind: complex128"),
		},
		{
			name:  "NoInput",
			input: &struct{ Foo complex128 }{},
			err:   errors.New("unsupported kind: complex128"),
		},
		{
			name: "Valid",
			input: fork{
				PreviousVersion: [4]byte{0x9f, 0x41, 0xbd, 0x5b},
				CurrentVersion:  [4]byte{0xcb, 0xb0, 0xf1, 0xd7},
				Epoch:           11971467576204192310,
			},
			output: [32]byte{0x3a, 0xd1, 0x26, 0x4c, 0x33, 0xbc, 0x66, 0xb4, 0x3a, 0x49, 0xb1, 0x25, 0x8b, 0x88, 0xf3, 0x4b, 0x8d, 0xbf, 0xa1, 0x64, 0x9f, 0x17, 0xe6, 0xdf, 0x55, 0x0f, 0x58, 0x96, 0x50, 0xd3, 0x49, 0x92},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := HashTreeRoot(test.input)
			if test.err == nil {
				if err != nil {
					t.Errorf("unexpected error %v", err)
				}
				if bytes.Compare(test.output[:], output[:]) != 0 {
					t.Errorf("incorrect output: expected %v; received %v", test.output, output)
				}
			} else {
				if err == nil {
					t.Fatalf("missing expected error %v", test.err)
				}
				if test.err.Error() != err.Error() {
					t.Errorf("incorrect error: expected %v; received %v", test.err, err)
				}
			}
		})
	}
}

func TestHashTreeRootBitlist(t *testing.T) {
	tests := []struct {
		name        string
		input       bitfield.Bitlist
		maxCapacity uint64
		output      []byte
		err         error
	}{
		{
			name:        "Nil",
			input:       nil,
			maxCapacity: 0,
			// Hash([]byte{})
			output: hexDecodeOrDie(t, "f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b"),
			err:    nil,
		},
		{
			name:        "SampleBitlist",
			input:       bitfield.Bitlist{1, 2, 3},
			maxCapacity: 4,
			// Known output hash.
			output: hexDecodeOrDie(t, "835e878350f244651619cbac69de3002251be60225ba0d6ac999b5becb469281"),
			err:    nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := HashTreeRootBitlist(test.input, test.maxCapacity)
			if test.err == nil {
				if err != nil {
					t.Fatalf("unexpected error %v", err)
				}
				if bytes.Compare(test.output[:], output[:]) != 0 {
					t.Errorf("incorrect output: expected %#x; received %#x", test.output, output)
				}
			} else {
				if err == nil {
					t.Fatalf("missing expected error %v", test.err)
				}
				if test.err.Error() != err.Error() {
					t.Errorf("incorrect error: expected %#x; received %#x", test.err, err)
				}
			}
		})
	}
}

func TestHashTreeRootWithCapacity(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		maxCapacity uint64
		output      [32]byte
		err         error
	}{
		{
			name: "Nil",
			err:  errors.New("untyped nil is not supported"),
		},
		{
			name:  "NotSlice",
			input: "foo",
			err:   errors.New("expected slice-kind input, received string"),
		},
		{
			name:  "InvalidSlice1",
			input: []complex128{complex(1, 1)},
			err:   errors.New("unsupported kind: complex128"),
		},
		{
			name:  "InvalidSlice2",
			input: []struct{ Foo complex128 }{{Foo: complex(1, 1)}},
			err:   errors.New("unsupported kind: complex128"),
		},
		{
			name:   "NoInput",
			input:  []uint32{},
			output: [32]byte{0xf5, 0xa5, 0xfd, 0x42, 0xd1, 0x6a, 0x20, 0x30, 0x27, 0x98, 0xef, 0x6e, 0xd3, 0x09, 0x97, 0x9b, 0x43, 0x00, 0x3d, 0x23, 0x20, 0xd9, 0xf0, 0xe8, 0xea, 0x98, 0x31, 0xa9, 0x27, 0x59, 0xfb, 0x4b},
		},
		{
			name: "Valid",
			input: []fork{{
				PreviousVersion: [4]byte{0x9f, 0x41, 0xbd, 0x5b},
				CurrentVersion:  [4]byte{0xcb, 0xb0, 0xf1, 0xd7},
				Epoch:           11971467576204192310,
			}},
			maxCapacity: 100,
			output:      [32]byte{0x5c, 0xa4, 0xd8, 0xbf, 0x17, 0xb9, 0x53, 0x6d, 0x69, 0x56, 0xee, 0x48, 0xfa, 0x3d, 0xc6, 0x91, 0xe3, 0x52, 0x48, 0xbd, 0x09, 0xb2, 0x9b, 0x1b, 0x5b, 0xa4, 0x5a, 0x0e, 0xd5, 0xda, 0xe0, 0xd9},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := HashTreeRootWithCapacity(test.input, test.maxCapacity)
			if test.err == nil {
				if err != nil {
					t.Fatalf("unexpected error %v", err)
				}
				if bytes.Compare(test.output[:], output[:]) != 0 {
					t.Errorf("incorrect output: expected %v; received %v", test.output, output)
				}
			} else {
				if err == nil {
					t.Fatalf("missing expected error %v", test.err)
				}
				if test.err.Error() != err.Error() {
					t.Errorf("incorrect error: expected %v; received %v", test.err, err)
				}
			}
		})
	}
}

func TestProtobufSSZFieldsIgnored(t *testing.T) {
	withProto := &simpleProtoMessage{
		Foo: []byte("foo"),
		Bar: 9001,
	}
	noProto := &simpleNonProtoMessage{
		Foo: []byte("foo"),
		Bar: 9001,
	}
	enc, err := Marshal(withProto)
	if err != nil {
		t.Fatal(err)
	}
	enc2, err := Marshal(noProto)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(enc, enc2) {
		t.Errorf("Wanted %v, received %v", enc, enc2)
	}
	withProtoDecoded := &simpleProtoMessage{}
	if err := Unmarshal(enc, withProtoDecoded); err != nil {
		t.Fatal(err)
	}
	noProtoDecoded := &simpleNonProtoMessage{}
	if err := Unmarshal(enc2, noProtoDecoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(noProto, noProtoDecoded) {
		t.Errorf("Wanted %v, received %v", noProto, noProtoDecoded)
	}
	if !reflect.DeepEqual(withProto, withProtoDecoded) {
		t.Errorf("Wanted %v, received %v", withProto, withProtoDecoded)
	}
}

func TestNilPointerHashTreeRoot(t *testing.T) {
	type nilItem struct {
		Field1 []*fork
		Field2 uint64
	}
	i := &nilItem{
		Field1: []*fork{nil},
		Field2: 10,
	}
	if _, err := HashTreeRoot(i); err != nil {
		t.Fatal(err)
	}
}

func TestNilInstantiationMarshalEquality(t *testing.T) {
	type exampleBody struct {
		Epoch uint64
	}
	type example struct {
		Slot uint64
		Root [32]byte
		Body *exampleBody
	}
	root := [32]byte{1, 2, 3, 4}
	item := &example{
		Slot: 5,
		Root: root,
		Body: nil,
	}
	item2 := &example{
		Slot: 5,
		Root: root,
		Body: &exampleBody{},
	}
	enc, err := Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	enc2, err := Marshal(item2)
	if err != nil {
		t.Fatal(err)
	}
	dec := &example{}
	if err := Unmarshal(enc, dec); err != nil {
		t.Fatal(err)
	}
	dec2 := &example{}
	if err := Unmarshal(enc2, dec2); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(enc, enc2) {
		t.Errorf("Unequal marshalings %v != %v", enc, enc2)
	}
}

func TestEmptyDataUnmarshal(t *testing.T) {
	msg := &simpleProtoMessage{}
	if err := Unmarshal([]byte{}, msg); err == nil {
		t.Error("Expected unmarshal to fail when attempting to unmarshal from an empty byte slice")
	}
}

func TestHashTreeRootWithCapacity_FailsWithNonSliceType(t *testing.T) {
	forkItem := fork{
		Epoch: 11971467576204192310,
	}
	capacity := uint64(100)
	if _, err := HashTreeRootWithCapacity(forkItem, capacity); err == nil {
		t.Error("Expected hash tree root to fail with non-slice type")
	}
}

func TestHashTreeRootWithCapacity_HashesCorrectly(t *testing.T) {
	capacity := uint64(1099511627776)
	balances := make([]uint64, 512)
	for i := 0; i < len(balances); i++ {
		balances[i] = 32000000000
	}
	root, err := HashTreeRootWithCapacity(balances, capacity)
	if err != nil {
		t.Fatal(err)
	}
	// Test case taken from validator balances of the state value in:
	// https://github.com/ethereum/eth2.0-spec-tests/blob/v0.8.0/tests/sanity/slots/sanity_slots_mainnet.yaml.
	want, err := hex.DecodeString("21a67313b0c6f988aac4fb6dd68686e1329243f7f6af21b722f6b83ca8fed9a8")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(root[:], want) {
		t.Errorf("Mismatched roots, wanted %#x == %#x", root, want)
	}
}

// Regression test for https://github.com/prysmaticlabs/go-ssz/issues/46.
func TestHashTreeRoot_EncodeSliceLengthCorrectly(t *testing.T) {
	type accountBalances struct {
		Balances []uint64 `ssz-max:"1099511627776"` // Large uint64 capacity.
	}
	acct := accountBalances{
		Balances: make([]uint64, 512),
	}
	for i := 0; i < len(acct.Balances); i++ {
		acct.Balances[i] = 32000000000
	}
	root, err := HashTreeRoot(acct)
	if err != nil {
		t.Fatal(err)
	}
	// Test case taken from validator balances of the state value in:
	// https://github.com/ethereum/eth2.0-spec-tests/blob/v0.8.0/tests/sanity/slots/sanity_slots_mainnet.yaml.
	want, err := hex.DecodeString("21a67313b0c6f988aac4fb6dd68686e1329243f7f6af21b722f6b83ca8fed9a8")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(root[:], want) {
		t.Errorf("Mismatched roots, wanted %#x == %#x", root, want)
	}
}

func TestHashTreeRoot_ConcurrentAccess(t *testing.T) {
	item := &truncateSignatureCase{
		Slot:              10,
		PreviousBlockRoot: []byte{'a', 'b'},
		Signature:         []byte("TESTING23"),
	}
	var wg sync.WaitGroup
	// We ensure the hash tree root function can be computed in a thread-safe manner.
	// No panic from this test is a successful run.
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func(tt *testing.T, w *sync.WaitGroup) {
			if _, err := HashTreeRoot(item); err != nil {
				tt.Fatal(err)
			}
			w.Done()
		}(t, &wg)
	}
	wg.Wait()
}

func TestSigningRoot(t *testing.T) {
	type signingRootTest struct {
		Val1 interface{}
		Val2 interface{}
		Err  error
	}
	type truncateLastCase struct {
		Slot           uint64
		StateRoot      []byte
		TruncatedField []byte
	}
	var signingRootTests = []signingRootTest{
		{
			Val1: &truncateSignatureCase{Slot: 20, Signature: []byte{'A', 'B'}},
			Val2: &truncateSignatureCase{Slot: 20, Signature: []byte("TESTING")},
		},
		{
			Val1: &truncateSignatureCase{
				Slot:              10,
				PreviousBlockRoot: []byte{'a', 'b'},
				Signature:         []byte("TESTINGDIFF")},
			Val2: &truncateSignatureCase{
				Slot:              10,
				PreviousBlockRoot: []byte{'a', 'b'},
				Signature:         []byte("TESTING23")},
		},
		{
			Val1: truncateSignatureCase{Slot: 50, Signature: []byte("THIS")},
			Val2: truncateSignatureCase{Slot: 50, Signature: []byte("DOESNT")},
		},
		{
			Val1: truncateSignatureCase{Signature: []byte("MATTER")},
			Val2: truncateSignatureCase{Signature: []byte("TESTING")},
		},
		{
			Val1: truncateLastCase{
				Slot:           5,
				StateRoot:      []byte("MATTERS"),
				TruncatedField: []byte("DOESNT MATTER"),
			},
			Val2: truncateLastCase{
				Slot:           5,
				StateRoot:      []byte("MATTERS"),
				TruncatedField: []byte("SHOULDNT MATTER"),
			},
		},
		{
			Val1: truncateLastCase{
				Slot:           550,
				StateRoot:      []byte("SHOULD"),
				TruncatedField: []byte("DOESNT"),
			},
			Val2: truncateLastCase{
				Slot:           550,
				StateRoot:      []byte("SHOULD"),
				TruncatedField: []byte("SHOULDNT"),
			},
		},
		{
			Val1: nil,
			Err:  errors.New("value cannot be nil"),
			Val2: nil,
		},
	}

	for i, test := range signingRootTests {
		output1, err := SigningRoot(test.Val1)
		if test.Err != nil {
			if err == nil {
				t.Fatalf("missing expected error of test %d value 1", i)
			}
			if test.Err.Error() != err.Error() {
				t.Fatalf("incorrect error at test %d value 1 %v", i, err)
			}
		} else {
			if err != nil {
				t.Fatalf("could not get the signing root of test %d, value 1 %v", i, err)
			}

			output2, err := SigningRoot(test.Val2)
			if err != nil {
				t.Errorf("could not get the signing root of test %d, value 2 %v", i, err)
			}
			// Check values have same result hash
			if !bytes.Equal(output1[:], output2[:]) {
				t.Errorf("test %d: hash mismatch: %X\n != %X", i, output1, output2)
			}
		}
	}
}

func TestSigningRoot_ConcurrentAccess(t *testing.T) {
	item := &truncateSignatureCase{
		Slot:              10,
		PreviousBlockRoot: []byte{'a', 'b'},
		Signature:         []byte("TESTING23"),
	}
	var wg sync.WaitGroup
	// We ensure the signing root function can be computed in a thread-safe manner.
	// No panic from this test is a successful run.
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func(tt *testing.T, w *sync.WaitGroup) {
			if _, err := SigningRoot(item); err != nil {
				tt.Fatal(err)
			}
			w.Done()
		}(t, &wg)
	}
	wg.Wait()
}

func hexDecodeOrDie(t *testing.T, s string) []byte {
	res, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return res
}
