package ssz

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func init() {
	useCache = true
}

type fork struct {
	PreviousVersion [4]byte
	CurrentVersion  [4]byte
	Epoch           uint64
}

type accountBalances struct {
	Balances []uint64 `ssz-max:"1099511627776"` // Large uint64 capacity.
}

func TestHashTreeRoot(t *testing.T) {
	useCache = false
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

// Regression test for https://github.com/prysmaticlabs/go-ssz/issues/46.
func TestHashTreeRoot_EncodeSliceLengthCorrectly(t *testing.T) {
	useCache = false
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
	useCache = true
}
