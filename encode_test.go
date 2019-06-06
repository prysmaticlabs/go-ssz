package ssz

import (
	"bytes"
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
		CurrentVersion:     []byte{203, 176, 241, 215},
		Epoch: 11971467576204192310,
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

	//cross := &crosslink{
	//	PreviousCrosslinkRoot: []byte{106, 119, 89, 97, 169, 156, 18, 151, 52, 60, 183, 231, 95, 0, 89, 103, 243,
	//		71, 219, 104, 221, 29, 176, 212, 140, 95, 246, 9, 140, 157, 77, 24},
	//	CrosslinkDataRoot:     []byte{75, 139, 131, 86, 243, 148, 75, 245, 90, 209, 33, 107, 104, 208, 144, 164, 137,
	//		94, 119, 36, 40, 152, 97, 144, 122, 245, 50, 247, 34, 155, 23, 89},
	//	Epoch: 11971467576204192310,
	//}
	//want = []byte{160, 178, 42, 13, 65, 67, 133, 92, 38, 11, 145, 55, 206, 161, 122, 235, 200, 23, 51, 157, 218, 89, 37,
	//	87, 117, 59, 103, 202, 58, 191, 241, 165}
	//buffer = new(bytes.Buffer)
	//if err := Encode(buffer, cross); err != nil {
	//	panic(err)
	//}
	//encodedBytes = buffer.Bytes()
	//if !bytes.Equal(want, encodedBytes) {
	//	t.Errorf("encode() = %v, want %v", want, encodedBytes)
	//}
}
