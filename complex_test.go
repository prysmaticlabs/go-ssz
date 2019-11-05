package ssz

import "testing"

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
	A uint16
	B []uint16 `ssz-max:"128"`
	C uint8
	D []byte `ssz-size:"256"`
	E varTestStruct
	F []fixedTestStruct `ssz-size:"4"`
	G []varTestStruct   `ssz-size:"2"`
}

func TestItem(t *testing.T) {
	someRoot := [256]byte{1, 2, 3, 4}
	fixedItem := fixedTestStruct{
		A: 1,
		B: 6,
		C: 7,
	}
	varItem := varTestStruct{
		A: 5,
		B: []uint16{6, 7, 8},
		C: 1,
	}
	fItems := [4]fixedTestStruct{fixedItem, fixedItem, fixedItem, fixedItem}
	varItems := [2]varTestStruct{varItem, varItem}
	c := complexTestStruct{
		A: 3,
		B: []uint16{1, 2, 3},
		C: 0,
		D: someRoot[:],
		E: varItem,
		F: fItems[:],
		G: varItems[:],
	}
	enc, _ := Marshal(c)
	var dec complexTestStruct
	Unmarshal(enc, &dec)
	t.Log(dec)
}
