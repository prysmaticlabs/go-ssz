package ssz

import (
	"bytes"
	"encoding/hex"
	"sync"
	"testing"
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

func TestHashTreeRoot(t *testing.T) {
	var currentVersion [4]byte
	var previousVersion [4]byte
	prev, err := hex.DecodeString("9f41bd5b")
	if err != nil {
		t.Fatal(err)
	}
	copy(previousVersion[:], prev)
	curr, err := hex.DecodeString("cbb0f1d7")
	if err != nil {
		t.Fatal(err)
	}
	copy(currentVersion[:], curr)
	forkItem := fork{
		PreviousVersion: previousVersion,
		CurrentVersion:  currentVersion,
		Epoch:           11971467576204192310,
	}
	root, err := HashTreeRoot(forkItem)
	if err != nil {
		t.Fatal(err)
	}
	want, err := hex.DecodeString("3ad1264c33bc66b43a49b1258b88f34b8dbfa1649f17e6df550f589650d34992")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(root[:], want) {
		t.Errorf("want %#x, HashTreeRoot() = %#x", want, root)
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
	}

	for i, test := range signingRootTests {
		output1, err := SigningRoot(test.Val1)
		if err != nil {
			t.Errorf("could not get the signing root of test %d, value 1 %v", i, err)
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
