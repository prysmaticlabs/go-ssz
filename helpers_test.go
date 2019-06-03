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
        	name: "an item having less than BytesPerChunk should return a padded chunk",
        	input: [][]byte{make([]byte, BytesPerChunk-4)},
			output: [][]byte{make([]byte, BytesPerChunk)},
		},
		{
			name: "two items having less than BytesPerChunk should return two chunks",
			input: [][]byte{make([]byte, BytesPerChunk-5), make([]byte, BytesPerChunk-5)},
			output: [][]byte{make([]byte, BytesPerChunk), make([]byte, BytesPerChunk)},
		},
		{
			name: "two items with length BytesPerChunk/2 should return one chunk",
			input: [][]byte{make([]byte, BytesPerChunk/2), make([]byte, BytesPerChunk/2)},
			output: [][]byte{make([]byte, BytesPerChunk)},
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
