package exp

import (
	"io/ioutil"
	"testing"

	"github.com/prysmaticlabs/go-ssz"
	"github.com/prysmaticlabs/go-ssz/types"
	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
)

func BenchmarkHashTreeRoot_Old(b *testing.B) {
	b.StopTimer()
	types.ToggleCache(false)
	enc, err := ioutil.ReadFile("genesis.ssz")
	if err != nil {
		panic(err)
	}
	st := &pb.BeaconState{}
	if err := ssz.Unmarshal(enc, st); err != nil {
		panic(err)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ssz.HashTreeRoot(st); err != nil {
			panic(err)
		}
	}
}

func BenchmarkHashTreeRoot_New(b *testing.B) {
	b.StopTimer()
	enc, err := ioutil.ReadFile("genesis.ssz")
	if err != nil {
		panic(err)
	}
	st := &pb.BeaconState{}
	if err := ssz.Unmarshal(enc, st); err != nil {
		panic(err)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		StateRoot(st)
	}
}
