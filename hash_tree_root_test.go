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

type Header struct {
	Slot       uint64
	ParentRoot []byte `ssz-size:"32"`
	StateRoot  []byte `ssz-size:"32"`
	BodyRoot   []byte `ssz-size:"32"`
	Signature  []byte `ssz-size:"96"`
}

type NoSignatureHeader struct {
	Slot       uint64
	ParentRoot []byte `ssz-size:"32"`
	StateRoot  []byte `ssz-size:"32"`
	BodyRoot   []byte `ssz-size:"32"`
}

func TestEdgeCase(t *testing.T) {
	h := &Header{
		Slot:       0,
		ParentRoot: make([]byte, 32),
		StateRoot:  []byte{3, 243, 60, 124, 153, 123, 57, 96, 95, 31, 255, 43, 95, 164, 219, 20, 5, 177, 147, 187, 150, 17, 32, 108, 197, 10, 251, 70, 9, 96, 253, 111},
		BodyRoot:   []byte{2, 33, 253, 156, 165, 71, 186, 33, 197, 248, 223, 7, 108, 127, 27, 130, 74, 234, 162, 8, 37, 60, 99, 224, 186, 108, 79, 109, 102, 157, 74, 91},
		Signature:  make([]byte, 96),
	}
	h2 := &Header{
		Slot:       0,
		ParentRoot: make([]byte, 32),
		StateRoot:  []byte{3, 243, 60, 124, 153, 123, 57, 96, 95, 31, 255, 43, 95, 164, 219, 20, 5, 177, 147, 187, 150, 17, 32, 108, 197, 10, 251, 70, 9, 96, 253, 111},
		BodyRoot:   []byte{2, 33, 253, 156, 165, 71, 186, 33, 197, 248, 223, 7, 108, 127, 27, 130, 74, 234, 162, 8, 37, 60, 99, 224, 186, 108, 79, 109, 102, 157, 74, 91},
	}
	result, err := SigningRoot(h)
	if err != nil {
		t.Fatal(err)
	}
	root, err := HashTreeRoot(h2)
	if err != nil {
		t.Fatal(err)
	}
	if root != result {
		t.Errorf("Mismatched roots HashTreeRoot(StructWithoutSignature) = %#x != SigningRoot(Struct) = %#x", root, result)
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
