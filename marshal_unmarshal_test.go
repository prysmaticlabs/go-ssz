package ssz

import (
	"fmt"
	"reflect"
	"testing"
)

type fork struct {
	PreviousVersion [4]byte
	CurrentVersion  [4]byte
	Epoch           uint64
}

type testFork struct {
	Epoch uint64
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

type Inner struct {
	C []byte
	D byte
}

type NestedItem struct {
	A byte
	B []Inner
	Z []Inner
}

func TestSpecVector(t *testing.T) {
	item := NestedItem{
		A: byte(1),
		B: []Inner{
			{
				C: []byte{2},
				D: byte(3),
			},
		},
		Z: []Inner{
			{
				C: []byte{4},
				D: byte(5),
			},
		},
	}
	encoded, err := Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	//ptr := new(NestedItem)
	//if err := Unmarshal(encoded, ptr); err != nil {
	//	t.Fatal(err)
	//}
	//if !reflect.DeepEqual(item, *ptr) {
	//	t.Errorf("Expected %v, received %v", item, *ptr)
	//}
}

func TestMarshalUnmarshal(t *testing.T) {
	forkExample := testFork{
		Epoch: 5,
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
		/*	// Bool test cases.
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
			{input: forkExample, ptr: new(fork)}, */
		//// Non-basic type slice/array test cases.
		{input: []testFork{forkExample, forkExample}, ptr: new([]testFork)},
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
