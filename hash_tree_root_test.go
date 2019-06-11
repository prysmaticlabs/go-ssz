package ssz

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestHashTreeRoot(t *testing.T) {
	fork := &fork{
		PreviousVersion: [4]byte{159, 65, 189, 91},
		CurrentVersion:  [4]byte{203, 176, 241, 215},
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

func BenchmarkHashTreeRoot(b *testing.B) {
	frk := &fork{
		PreviousVersion: [4]byte{159, 65, 189, 91},
		CurrentVersion:  [4]byte{203, 176, 241, 215},
		Epoch:           11971467576204192310,
	}
	for n := 0; n < b.N; n++ {
		if _, err := HashTreeRoot(frk); err != nil {
			b.Fatal(err)
		}
	}
}
