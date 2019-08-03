package types

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type compactCommittee struct {
	Indices    [10]uint64
	Slashed    [10]uint64
	BlockRoots [64][32]byte
}

func TestBasicSSZ_Root(t *testing.T) {
	var roots [64][32]byte
	for i := 0; i < 64; i++ {
		enc := fmt.Sprintf("%d", i)
		r := hash([]byte(enc))
		roots[i] = r
	}
	item := compactCommittee{
		Indices:    [10]uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Slashed:    [10]uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		BlockRoots: roots,
	}
	val := reflect.ValueOf(item)
	typ := reflect.TypeOf(item)
	factory, _ := SSZFactory(val, typ)
	got, err := factory.Root(val, typ, 0)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	got, err = factory.Root(val, typ, 0)
	if err != nil {
		t.Fatal(err)
	}
	end := time.Now()
	fmt.Printf("%#x and time %v\n", got, end.Sub(start))
	start = time.Now()
	got, err = factory.Root(val, typ, 0)
	if err != nil {
		t.Fatal(err)
	}
	end = time.Now()
	fmt.Printf("%#x and time %v\n", got, end.Sub(start))
}
