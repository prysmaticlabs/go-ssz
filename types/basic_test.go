package types

import (
	"fmt"
	"reflect"
	"testing"
)

type compactCommittee struct {
	BlockRoots [][]byte `ssz-size:"64,32"`
}

func TestBasicSSZ_Root(t *testing.T) {
	roots := make([][]byte, 64)
	for i := 0; i < 64; i++ {
		enc := fmt.Sprintf("%d", i)
		r := hash([]byte(enc))
		roots[i] = r[:]
	}
	item := compactCommittee{
		BlockRoots: roots,
	}
	val := reflect.ValueOf(item)
	typ := reflect.TypeOf(item)
	factory, _ := SSZFactory(val, typ)
	if _, err := factory.Root(val, typ, 0); err != nil {
		t.Fatal(err)
	}
	//start := time.Now()
	//got, err = factory.Root(val, typ, 0)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//end := time.Now()
	//fmt.Printf("%#x and time %v\n", got, end.Sub(start))
	//start = time.Now()
	//got, err = factory.Root(val, typ, 0)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//end = time.Now()
	//fmt.Printf("%#x and time %v\n", got, end.Sub(start))
}
