package ssz

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
)

type fork struct {
	PreviousVersion [4]byte
	CurrentVersion  [4]byte
	Epoch           uint64
}

type nestedItem struct {
	Field1 []uint64
	Field2 *fork
	Field3 [3]byte
}

type prysmState struct {
	HeadRoot []uint64 `ssz:"size=32"`
	ForkType []byte   `ssz:"size=4"`
	Epoch    uint64
}

type Crosslink struct {
	Shard      uint64
	StartEpoch uint64
	EndEpoch   uint64
	ParentRoot [32]byte
	DataRoot   [32]byte
}

type AttestationData struct {
	BeaconBlockRoot [32]byte
	SourceEpoch     uint64
	SourceRoot      [32]byte
	TargetEpoch     uint64
	TargetRoot      [32]byte
	Crosslink       Crosslink
}

type Attestation struct {
	AggregationBitfield []byte
	Data                AttestationData
	CustodyBitfield     []byte
	Signature           [96]byte
}

func TestSpecVector(t *testing.T) {
	exampleAttestation := Attestation{
		AggregationBitfield: []byte{159},
		Data: AttestationData{
			BeaconBlockRoot: [32]byte{65, 189, 91, 203, 176, 241, 215, 189, 166, 236, 135, 7, 215, 119, 198, 241, 63, 166, 13, 230, 40, 28, 95, 120, 222, 63, 97, 139, 26, 146, 63, 3},
			SourceEpoch:     3997959117937236768,
			SourceRoot:      [32]byte{71, 46, 234, 222, 196, 99, 40, 195, 204, 18, 35, 158, 158, 113, 32, 33, 0, 248, 223, 1, 53, 198, 55, 245, 251, 42, 223, 42, 74, 80, 246, 50},
			TargetEpoch:     3777515321107143329,
			TargetRoot:      [32]byte{240, 247, 176, 50, 247, 247, 228, 98, 76, 5, 92, 106, 42, 239, 37, 67, 16, 84, 77, 209, 154, 150, 0, 152, 173, 181, 86, 16, 79, 90, 209, 78},
			Crosslink: Crosslink{
				Shard:      12846677991095410117,
				StartEpoch: 8876912483467349126,
				EndEpoch:   3248842131919680082,
				ParentRoot: [32]byte{65, 249, 240, 5, 243, 191, 91, 216, 103, 4, 140, 201, 107, 93, 96, 148, 215, 2, 115, 11, 181, 46, 159, 244, 244, 50, 30, 193, 62, 237, 209, 241},
				DataRoot:   [32]byte{118, 88, 131, 90, 228, 134, 64, 198, 118, 27, 150, 191, 199, 204, 94, 220, 75, 9, 110, 242, 250, 23, 53, 87, 131, 156, 92, 235, 37, 87, 70, 236},
			},
		},
		CustodyBitfield: []byte{179},
		Signature:       [96]byte{139, 23, 79, 175, 81, 78, 45, 204, 20, 160, 38, 184, 176, 79, 255, 123, 41, 184, 12, 20, 252, 153, 136, 236, 103, 8, 60, 189, 152, 88, 211, 64, 116, 166, 107, 37, 14, 250, 234, 163, 8, 205, 126, 85, 214, 53, 33, 251, 187, 144, 243, 33, 161, 229, 201, 105, 27, 43, 111, 95, 38, 15, 216, 107, 75, 36, 116, 236, 255, 158, 218, 43, 249, 133, 116, 245, 124, 176, 186, 81, 122, 70, 74, 120, 103, 230, 37, 28, 96, 210, 253, 136, 245, 45, 160, 239},
	}
	encoded, err := Marshal(exampleAttestation)
	if err != nil {
		t.Fatal(err)
	}
	ptr := new(Attestation)
	if err := Unmarshal(encoded, ptr); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exampleAttestation, *ptr) {
		t.Errorf("Expected %v, received %v", exampleAttestation, *ptr)
	}
	expect := []byte{48, 1, 0, 0, 65, 189, 91, 203, 176, 241, 215, 189, 166, 236, 135, 7, 215, 119, 198, 241, 63, 166, 13, 230, 40, 28, 95, 120, 222, 63, 97, 139, 26, 146, 63, 3, 32, 219, 46, 187, 162, 154, 123, 55, 71, 46, 234, 222, 196, 99, 40, 195, 204, 18, 35, 158, 158, 113, 32, 33, 0, 248, 223, 1, 53, 198, 55, 245, 251, 42, 223, 42, 74, 80, 246, 50, 161, 218, 47, 160, 43, 110, 108, 52, 240, 247, 176, 50, 247, 247, 228, 98, 76, 5, 92, 106, 42, 239, 37, 67, 16, 84, 77, 209, 154, 150, 0, 152, 173, 181, 86, 16, 79, 90, 209, 78, 197, 29, 18, 123, 145, 145, 72, 178, 134, 76, 77, 47, 223, 32, 49, 123, 82, 50, 101, 180, 180, 52, 22, 45, 65, 249, 240, 5, 243, 191, 91, 216, 103, 4, 140, 201, 107, 93, 96, 148, 215, 2, 115, 11, 181, 46, 159, 244, 244, 50, 30, 193, 62, 237, 209, 241, 118, 88, 131, 90, 228, 134, 64, 198, 118, 27, 150, 191, 199, 204, 94, 220, 75, 9, 110, 242, 250, 23, 53, 87, 131, 156, 92, 235, 37, 87, 70, 236, 49, 1, 0, 0, 139, 23, 79, 175, 81, 78, 45, 204, 20, 160, 38, 184, 176, 79, 255, 123, 41, 184, 12, 20, 252, 153, 136, 236, 103, 8, 60, 189, 152, 88, 211, 64, 116, 166, 107, 37, 14, 250, 234, 163, 8, 205, 126, 85, 214, 53, 33, 251, 187, 144, 243, 33, 161, 229, 201, 105, 27, 43, 111, 95, 38, 15, 216, 107, 75, 36, 116, 236, 255, 158, 218, 43, 249, 133, 116, 245, 124, 176, 186, 81, 122, 70, 74, 120, 103, 230, 37, 28, 96, 210, 253, 136, 245, 45, 160, 239, 159, 179}
	if len(encoded) != len(expect) {
		t.Fatalf("Expected encoded.length == %d, received %d", len(expect), len(encoded))
	}
	if !bytes.Equal(encoded, expect) {
		t.Fatalf("Expected %#x, received %#x", expect, encoded)
	}
	expectedRoot, err := hex.DecodeString("41ab56baf36308c1db1ca3518ae5b1b783d749a9165ecc711c39fdee9ee95742")
	if err != nil {
		t.Fatal(err)
	}
	root, err := HashTreeRoot(exampleAttestation)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expectedRoot, root[:]) {
		t.Fatalf("Expected %#x, received %#x", expectedRoot, root)
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	forkExample := fork{
		PreviousVersion: [4]byte{2, 3, 4, 1},
		CurrentVersion:  [4]byte{5, 6, 7, 8},
		Epoch:           5,
	}
	//nestedItemExample := nestedItem{
	//	Field1: []uint64{1, 2, 3, 4},
	//	Field2: &forkExample,
	//	Field3: [3]byte{32, 33, 34},
	//}
	//headRoot := [32]uint64{3, 4, 5}
	//forkType := [4]byte{6, 7}
	//stateExample := prysmState{
	//	HeadRoot: headRoot[:],
	//	ForkType: forkType[:],
	//	Epoch:    5,
	//}
	tests := []struct {
		input interface{}
		ptr   interface{}
	}{
		// Bool test cases.
		{input: true, ptr: new(bool)},
		{input: false, ptr: new(bool)},
		// Uint8 test cases.
		{input: byte(1), ptr: new(byte)},
		{input: byte(0), ptr: new(byte)},
		// Uint16 test cases.
		{input: uint16(100), ptr: new(uint16)},
		{input: uint16(232), ptr: new(uint16)},
		// Uint32 test cases.
		{input: uint32(1), ptr: new(uint32)},
		{input: uint32(1029391), ptr: new(uint32)},
		// Uint64 test cases.
		{input: uint64(5), ptr: new(uint64)},
		{input: uint64(23929309), ptr: new(uint64)},
		// Byte slice, byte array test cases.
		{input: [8]byte{1, 2, 3, 4, 5, 6, 7, 8}, ptr: new([8]byte)},
		{input: []byte{9, 8, 9, 8}, ptr: new([]byte)},
		// Basic type array test cases.
		{input: [12]uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, ptr: new([12]uint64)},
		{input: [100]bool{true, false, true, true}, ptr: new([100]bool)},
		{input: [20]uint16{3, 4, 5}, ptr: new([20]uint16)},
		{input: [20]uint32{4, 5}, ptr: new([20]uint32)},
		{input: [20][2]uint32{{3, 4}, {5}, {8}, {9, 10}}, ptr: new([20][2]uint32)},
		// Basic type slice test cases.
		{input: []uint64{1, 2, 3}, ptr: new([]uint64)},
		{input: []bool{true, false, true, true, true}, ptr: new([]bool)},
		{input: []uint32{0, 0, 0}, ptr: new([]uint32)},
		{input: []uint32{92939, 232, 222}, ptr: new([]uint32)},
		// Struct decoding test cases.
		{input: forkExample, ptr: new(fork)},
		//// Non-basic type slice/array test cases.
		{input: []fork{forkExample, forkExample}, ptr: new([]fork)},
		//{input: [][]uint64{{4, 3, 2}, {1}, {0}}, ptr: new([][]uint64)},
		//{input: [][][]uint64{{{1, 2}, {3}}, {{4, 5}}, {{0}}}, ptr: new([][][]uint64)},
		//{input: [][3]uint64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, ptr: new([][3]uint64)},
		//{input: [3][]uint64{{1, 2}, {4, 5, 6}, {7}}, ptr: new([3][]uint64)},
		//{input: [][4]fork{{forkExample, forkExample, forkExample}}, ptr: new([][4]fork)},
		// Pointer-type test cases.
		//{input: &forkExample, ptr: new(fork)},
		//{input: &nestedItemExample, ptr: new(nestedItem)},
		//{input: []*fork{&forkExample, &forkExample}, ptr: new([]*fork)},
		//{input: []*nestedItem{&nestedItemExample, &nestedItemExample}, ptr: new([]*nestedItem)},
		//{input: [2]*nestedItem{&nestedItemExample, &nestedItemExample}, ptr: new([2]*nestedItem)},
		//{input: [2]*fork{&forkExample, &forkExample}, ptr: new([2]*fork)},
		//{input: stateExample, ptr: new(prysmState)},
	}
	for _, tt := range tests {
		serializedItem, err := Marshal(tt.input)
		if err != nil {
			panic(err)
		}
		fmt.Println("---Serialized")
		fmt.Println(serializedItem)
		if err := Unmarshal(serializedItem, tt.ptr); err != nil {
			t.Fatal(err)
		}
		output := reflect.ValueOf(tt.ptr)
		inputVal := reflect.ValueOf(tt.input)
		if inputVal.Kind() == reflect.Ptr {
			fmt.Println(output.Interface())
			if !reflect.DeepEqual(output.Interface(), tt.input) {
				t.Errorf("Expected %d, received %d", tt.input, output.Interface())
			}
		} else {
			fmt.Println(output.Elem().Interface())
			if !reflect.DeepEqual(output.Elem().Interface(), tt.input) {
				t.Errorf("Expected %d, received %d", tt.input, output.Elem().Interface())
			}
		}
	}
}
