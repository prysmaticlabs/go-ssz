package types

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type fork struct {
	PreviousVersion [4]uint16
	CurrentVersion  [4]uint16
	Epoch           uint64
}

func TestBasicSSZ_Root(t *testing.T) {
	forkItem := fork{
		PreviousVersion: [4]uint16{1, 2, 3, 4},
		CurrentVersion:  [4]uint16{1, 2, 3, 4},
		Epoch:           200,
	}
	val := reflect.ValueOf(forkItem)
	typ := reflect.TypeOf(forkItem)
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
