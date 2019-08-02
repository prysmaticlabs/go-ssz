package types

import (
	"bytes"
	"encoding/hex"
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
	got, err := factory.Root(val, typ)
	if err != nil {
		t.Fatal(err)
	}
	want, err := hex.DecodeString("3ad1264c33bc66b43a49b1258b88f34b8dbfa1649f17e6df550f589650d34992")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got[:], want) {
		t.Errorf("want %#x, HashTreeRoot() = %#x", want, got)
	}
	start := time.Now()
	got, err = factory.Root(val, typ)
	if err != nil {
		t.Fatal(err)
	}
	end := time.Now()
	fmt.Printf("%#x and time %v\n", got, end.Sub(start))
	start = time.Now()
	got, err = factory.Root(val, typ)
	if err != nil {
		t.Fatal(err)
	}
	end = time.Now()
	fmt.Printf("%#x and time %v\n", got, end.Sub(start))
}
