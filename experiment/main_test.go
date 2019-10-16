package main

import (
	"testing"

	"github.com/karlseguin/ccache"

	"github.com/prysmaticlabs/go-ssz"
	pb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
)

func BenchmarkOldMarshal(b *testing.B) {
	b.StopTimer()
	check := pb.Checkpoint{
		Epoch: 1,
		Root: make([]byte, 32),
	}
	copy(check.Root, "12345678910")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ssz.Marshal(check); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewMarshal(b *testing.B) {
	b.StopTimer()
	check := pb.Checkpoint{
		Epoch: 1,
		Root: make([]byte, 32),
	}
	copy(check.Root, "12345678910")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if _, err := fastMarshal(check); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewArrayHasher(b *testing.B) {
	b.StopTimer()
	bs := BeaconState{
		BlockRoots: [65536][32]byte{},
	}
	hasher := structHasher{
		hashCache:  ccache.New(ccache.Configure().MaxSize(10000)),
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if _, err := hasher.fastHashTreeRoot(bs); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOldArrayHasher(b *testing.B) {
	b.StopTimer()
	bs := BeaconState{
		BlockRoots: [65536][32]byte{},
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ssz.HashTreeRoot(bs); err != nil {
			b.Fatal(err)
		}
	}
}
