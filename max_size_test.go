package ssz

import (
	"strconv"
	"testing"
)

func TestMaxSize_Limit(t *testing.T) {
	type ex struct {
		Roots [][]byte `ssz-size:"?,32" ssz-max:"128"`
	}
	elem := &ex{
		Roots: make([][]byte, 0),
	}
	for i := 0; i < 129; i++ {
		r := make([]byte, 32)
		copy(r, strconv.Itoa(i))
		elem.Roots = append(elem.Roots, r)
	}
	if _, err := HashTreeRoot(elem); err != nil {
		t.Fatal(err)
	}
}
