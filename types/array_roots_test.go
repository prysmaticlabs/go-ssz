package types

import (
	"reflect"
	"testing"
)

type beaconState struct {
	BlockRoots [65536][32]byte
}

func BenchmarkRootsArray_Root_WithCache(b *testing.B) {
	b.StopTimer()
	bs := beaconState{
		BlockRoots: [65536][32]byte{},
	}
	for i := 0; i < len(bs.BlockRoots); i++ {
		bs.BlockRoots[i] = [32]byte{1, 2, 3}
	}
	ss := newRootsArraySSZ()
	v := reflect.ValueOf(bs.BlockRoots)
	typ := v.Type()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ss.Root(v, typ, "BlockRoots", 0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRootsArray_Root_MinimalChanges(b *testing.B) {
	b.StopTimer()
	bs := beaconState{
		BlockRoots: [65536][32]byte{},
	}
	for i := 0; i < len(bs.BlockRoots); i++ {
		bs.BlockRoots[i] = [32]byte{1, 2, 3}
	}
	ss := newRootsArraySSZ()
	v := reflect.ValueOf(bs.BlockRoots)
	typ := v.Type()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		bs.BlockRoots[i%len(bs.BlockRoots)] = [32]byte{4, 5, 6}
		if _, err := ss.Root(v, typ, "BlockRoots", 0); err != nil {
			b.Fatal(err)
		}
	}
}
