package ssz

import (
	"reflect"
	"testing"
)

func TestPack_NoItems(t *testing.T) {
	output, err := pack([][]byte{})
	if err != nil {
		t.Fatalf("pack() error = %v", err)
	}
	if len(output[0]) != BytesPerChunk {
		t.Errorf("Expected empty input to return an empty chunk, received %v", output)
	}
}

func TestPack_ExactBytePerChunkLength(t *testing.T) {
	input := [][]byte{}
	for i := 0; i < 10; i++ {
		item := make([]byte, BytesPerChunk)
		input = append(input, item)
	}
	output, err := pack(input)
	if err != nil {
		t.Fatalf("pack() error = %v", err)
	}
	if len(output) != 10 {
		t.Errorf("Expected empty input to return an empty chunk, received %v", output)
	}
	if !reflect.DeepEqual(output, input) {
		t.Errorf("pack() = %v, want %v", output, input)
	}
}

func TestPack_OK(t *testing.T) {
	tests := []struct {
		name    string
		input    [][]byte
		output    [][]byte
	}{
        {
        	name: "no items should return an empty chunk",
        	input: [][]byte{},
        	output: make([][]byte, BytesPerChunk),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pack(tt.input)
			if err != nil {
				t.Fatalf("pack() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("pack() = %v, want %v", got, tt.output)
			}
		})
	}
}
