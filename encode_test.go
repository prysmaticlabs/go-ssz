package ssz

import (
	"bytes"
	"encoding/hex"
	"testing"
)

type crosslink struct {
	Epoch                 uint64
	PreviousCrosslinkRoot []byte
	CrosslinkDataRoot     []byte
}

type fork struct {
	PreviousVersion []byte
	CurrentVersion  []byte
	Epoch           uint64
}

func TestEncode(t *testing.T) {
	fork := &fork{
		PreviousVersion: []byte{159, 65, 189, 91},
		CurrentVersion:  []byte{203, 176, 241, 215},
		Epoch:           11971467576204192310,
	}
	want := []byte{159, 65, 189, 91, 203, 176, 241, 215, 54, 234, 193, 63, 85, 50, 35, 166}
	buffer := new(bytes.Buffer)
	if err := Encode(buffer, fork); err != nil {
		panic(err)
	}
	encodedBytes := buffer.Bytes()
	if !bytes.Equal(want, encodedBytes) {
		t.Errorf("encode() = %v, want %v", encodedBytes, want)
	}
	prevRoot, err := hex.DecodeString("e8933c7bb4e15a6476373346d2334d8f845bc3c0c93d5d5acf3fd0fba9d7e8d9")
	if err != nil {
		t.Fatal(err)
	}
	root, err := hex.DecodeString("0f9e7e66592424d43d7d6109182b6519c0b748e6eb33cbccc1527aae78dc889f")
	if err != nil {
		t.Fatal(err)
	}
	cross := &crosslink{
		PreviousCrosslinkRoot: prevRoot,
		CrosslinkDataRoot:     root,
		Epoch:                 19993510755097755,
	}
	want, _ = hex.DecodeString("9bdc5efafd074700e8933c7bb4e15a6476373346d2334d8f845bc3c0c93d5d5acf3fd0fba9d7e8d90f9e7e66592424d43d7d6109182b6519c0b748e6eb33cbccc1527aae78dc889f")
	buffer = new(bytes.Buffer)
	if err := Encode(buffer, cross); err != nil {
		panic(err)
	}
	encodedBytes = buffer.Bytes()
	if !bytes.Equal(want, encodedBytes) {
		t.Errorf("want %v, encode() = %v", want, encodedBytes)
	}
}
