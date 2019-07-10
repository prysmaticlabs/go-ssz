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

type balanceContainer struct {
	Balances []uint64 `ssz-max:"1099511627776"`
}

func TestBalanceStruct(t *testing.T) {
	b := balanceContainer{
		Balances: make([]uint64, 512),
	}
	for i := 0; i < len(b.Balances); i++ {
		b.Balances[i] = uint64(32000000000)
	}
	root, err := HashTreeRoot(b)
	if err != nil {
		t.Fatal(err)
	}
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
