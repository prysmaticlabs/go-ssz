package ssz

import (
	"fmt"
	"testing"
)

type fixedTestStruct struct {
	A uint8
	B uint64
	C uint32
}

type varTestStruct struct {
	A uint16
	B []uint16 `ssz-max:"1024"`
	C uint8
}

type complexTestStruct struct {
	A uint16            // size = 2
	B []uint16          `ssz-max:"128"` // size = 2 + 4 = 6
	C uint8             // size = 1
	D []byte            `ssz-size:"256"` // size = 256
	E varTestStruct     // size 2 + 6 + 1 = 9 + 4 = 13
	F []fixedTestStruct `ssz-size:"4"` // size 13 * 4 = 52
	G []varTestStruct   `ssz-size:"2"` // size 14 + 4 =  18
}

func TestItem(t *testing.T) {
	input := hexDecodeOrDie(t, "0fbf47000000df490000004a000000dffce2db220f0e14dc7f85b33deecebead2cc9d8007282d99b0cce0d500798db0e6c269c2f4f5815b13fa24b121086729a4c1fd253000000a224583e8b07000000c261b00800000011000000328a07000000e2958bf3f6070000006a9374")
	fmt.Println(input)
	var dec complexTestStruct
	Unmarshal(input, &dec)
	t.Log(dec)
}
