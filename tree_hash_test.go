package ssz

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestTreeHash(t *testing.T) {
	fork := &fork{
		PreviousVersion: []byte{159, 65, 189, 91},
		CurrentVersion:  []byte{203, 176, 241, 215},
		Epoch:           11971467576204192310,
	}
	want, err := hex.DecodeString("3ad1264c33bc66b43a49b1258b88f34b8dbfa1649f17e6df550f589650d34992")
	if err != nil {
		t.Fatal(err)
	}
	root, err := HashTreeRoot(fork)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(root[:], want) {
		t.Errorf("want %v, HashTreeRoot() = %v", want, root)
	}
}
