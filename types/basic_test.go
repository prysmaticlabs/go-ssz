package types

import (
	"fmt"
	"reflect"
	"testing"
)

type depData struct {
	Pubkey     []byte `ssz-size:"48"`
	Withdrawal []byte `ssz-size:"32"`
	Amount     uint64
	Signature  []byte `ssz-size:"96"`
}

type deposit struct {
	Proof [][]byte `ssz-size:"33,32"`
	Data  depData
}

func TestBasicSSZ_Root(t *testing.T) {
	roots := make([][]byte, 33)
	emptyRoot := make([]byte, 32)
	emptyKey := make([]byte, 48)
	emptySig := make([]byte, 96)
	for i := 0; i < 32; i++ {
		roots[i] = emptyRoot
	}
	item := deposit{
		Proof: roots,
		Data: depData{
			Pubkey:     emptyKey,
			Withdrawal: emptyRoot,
			Amount:     0,
			Signature:  emptySig,
		},
	}
	val := reflect.ValueOf(item)
	typ := reflect.TypeOf(item)
	factory, _ := SSZFactory(val, typ)
	res, err := factory.Root(val, typ, 0)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#x\n", res)
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
