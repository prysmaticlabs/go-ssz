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

var exampleFork = fork{
	PreviousVersion: [4]byte{159, 65, 189, 91},
	CurrentVersion:  [4]byte{203, 176, 241, 215},
	Epoch:           11971467576204192310,
}

func TestHashTreeRoot(t *testing.T) {
	useCache = false
	want, err := hex.DecodeString("3ad1264c33bc66b43a49b1258b88f34b8dbfa1649f17e6df550f589650d34992")
	if err != nil {
		t.Fatal(err)
	}
	root, err := HashTreeRoot(exampleFork)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(root[:], want) {
		t.Errorf("want %v, HashTreeRoot() = %v", want, root)
	}
}

func TestHashTreeRootCached(t *testing.T) {
	useCache = false
	want, err := hex.DecodeString("3ad1264c33bc66b43a49b1258b88f34b8dbfa1649f17e6df550f589650d34992")
	if err != nil {
		t.Fatal(err)
	}

	root, err := HashTreeRoot(exampleFork)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(root[:], want) {
		t.Errorf("want %v, HashTreeRoot() = %v", want, root)
	}
	root, err = HashTreeRoot(exampleFork)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(root[:], want) {
		t.Errorf("want %v, HashTreeRoot() = %v", want, root)
	}
}

func BenchmarkHashTreeRoot(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if _, err := HashTreeRoot(exampleFork); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHashTreeRootCached(b *testing.B) {
	useCache = true
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if _, err := HashTreeRoot(exampleFork); err != nil {
			b.Fatal(err)
		}
	}
}
