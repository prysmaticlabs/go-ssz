package main

import (
	"log"
	"time"

	"github.com/prysmaticlabs/go-ssz"
)

type BeaconState struct {
	BlockRoots       [][]byte `ssz-size:"65536,32"`
	ActiveIndexRoots [][]byte `ssz-size:"2048,32"`
}

func main() {
	bs := BeaconState{
		BlockRoots:       make([][]byte, 65536),
		ActiveIndexRoots: make([][]byte, 2048),
	}
	for i := 0; i < len(bs.BlockRoots); i++ {
		item := [32]byte{1, 2, 3}
		bs.BlockRoots[i] = item[:]
	}
	for i := 0; i < len(bs.ActiveIndexRoots); i++ {
		item := [32]byte{4, 5, 6}
		bs.ActiveIndexRoots[i] = item[:]
	}
	start := time.Now()
	enc, err := ssz.HashTreeRoot(bs)
	if err != nil {
		panic(err)
	}
	end := time.Now()
	log.Printf("%#x in %v", enc, end.Sub(start))
	start = time.Now()
	enc, err = ssz.HashTreeRoot(bs)
	if err != nil {
		panic(err)
	}
	end = time.Now()
	log.Printf("%#x in %v", enc, end.Sub(start))
	// Change 2 elements.
	newItem1 := [32]byte{4, 5, 6}
	newItem2 := [32]byte{7, 8, 9}
	bs.BlockRoots[len(bs.BlockRoots)-1] = newItem1[:]
	bs.BlockRoots[1] = newItem2[:]
	// Change 1 element.
	newItem := [32]byte{1, 1, 1}
	bs.ActiveIndexRoots[1] = newItem[:]
	start = time.Now()
	enc, err = ssz.HashTreeRoot(bs)
	if err != nil {
		panic(err)
	}
	end = time.Now()
	log.Printf("%#x in %v", enc, end.Sub(start))
}
