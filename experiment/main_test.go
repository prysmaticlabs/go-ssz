package main

import (
	"strconv"
	"testing"

	"github.com/prysmaticlabs/go-ssz"
	"github.com/prysmaticlabs/go-ssz/types"
)

func BenchmarkSSZ_NoCache(b *testing.B) {
	b.StopTimer()
	types.ToggleCache(false)
	bs := &beaconState{
		BlockRoots: make([][]byte, 65536),
	}
	for i := 0; i < len(bs.BlockRoots); i++ {
		newItem := [32]byte{1, 2, 3}
		bs.BlockRoots[i] = newItem[:]
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ssz.HashTreeRoot(bs); err != nil {
			b.Fatal(err)
		}
	}
	types.ToggleCache(true)
}

func BenchmarkSSZ_WithCache(b *testing.B) {
	b.StopTimer()
	types.ToggleCache(true)
	bs := &beaconState{
		BlockRoots: make([][]byte, 65536),
	}
	for i := 0; i < len(bs.BlockRoots); i++ {
		newItem := [32]byte{1, 2, 3}
		bs.BlockRoots[i] = newItem[:]
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ssz.HashTreeRoot(bs); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSSZ_SingleElementChanged(b *testing.B) {
	b.StopTimer()
	types.ToggleCache(true)
	bs := &beaconState{
		BlockRoots: make([][]byte, 65536),
	}
	for i := 0; i < len(bs.BlockRoots); i++ {
		newItem := [32]byte{1, 2, 3}
		bs.BlockRoots[i] = newItem[:]
	}
	if _, err := ssz.HashTreeRoot(bs); err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		newItem := []byte(strconv.Itoa(i))
		newRoot := toBytes32(newItem)
		bs.BlockRoots[i%len(bs.BlockRoots)] = newRoot[:]
		if _, err := ssz.HashTreeRoot(bs); err != nil {
			b.Fatal(err)
		}
	}
}

func toBytes32(x []byte) [32]byte {
	var y [32]byte
	copy(y[:], x)
	return y
}
