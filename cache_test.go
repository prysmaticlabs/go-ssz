package ssz

import (
	"testing"
	"time"
)

type cachedItem struct {
	Slot            uint64
	CurrentVersion  [4]byte
	PreviousVersion [4]byte
	HistoricalRoots [][]byte `ssz-size:"?,32" ssz-max:"100"`
}

func TestNewCache(t *testing.T) {
	useCache = true
	i := &cachedItem{
		Slot:            200,
		CurrentVersion:  [4]byte{1, 2, 3, 4},
		PreviousVersion: [4]byte{5, 6, 7, 8},
		HistoricalRoots: [][]byte{
			make([]byte, 32),
			make([]byte, 32),
			make([]byte, 32),
			make([]byte, 32),
		},
	}
	start := time.Now()
	if _, err := HashTreeRoot(i); err != nil {
		t.Fatal(err)
	}
	end := time.Now()
	t.Logf("Took %v", end.Sub(start))
	start = time.Now()
	if _, err := HashTreeRoot(i); err != nil {
		t.Fatal(err)
	}
	end = time.Now()
	t.Logf("Took %v", end.Sub(start))
	useCache = true
}
