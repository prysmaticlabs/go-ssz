package main

import (
	"log"
	"time"

	"github.com/prysmaticlabs/go-ssz"
)

type beaconState struct {
	BlockRoots [][]byte `ssz-size:"65536,32"`
}

func main() {
	bs := &beaconState{
		BlockRoots: make([][]byte, 65536),
	}
	for i := 0; i < len(bs.BlockRoots); i++ {
		newItem := [32]byte{1, 2, 3}
		bs.BlockRoots[i] = newItem[:]
	}
	start := time.Now()
	if _, err := ssz.HashTreeRoot(bs); err != nil {
		panic(err)
	}
	end := time.Now()
	log.Printf("%v", end.Sub(start))
	start = time.Now()
	if _, err := ssz.HashTreeRoot(bs); err != nil {
		panic(err)
	}
	end = time.Now()
	log.Printf("%v", end.Sub(start))
	newRoot := [32]byte{4, 5, 6}
	bs.BlockRoots[2] = newRoot[:]
	start = time.Now()
	if _, err := ssz.HashTreeRoot(bs); err != nil {
		panic(err)
	}
	end = time.Now()
	log.Printf("%v", end.Sub(start))
}
