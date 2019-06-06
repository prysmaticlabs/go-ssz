package ssz

import (
	"bytes"
	"testing"
)

type crosslink struct {
	Epoch uint64
	PreviousCrosslinkRoot []byte
	CrosslinkDataRoot []byte
}

func TestEncode(t *testing.T) {
	tests := []struct {
		name   string
		input  *crosslink
		output []byte
	}{
		{
			name:   "basic crosslink",
			input:  &crosslink{

			},
			output: []byte{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := new(bytes.Buffer)
			if err := Encode(buffer, tt.input); err != nil {
				panic(err)
			}
			encodedBytes := buffer.Bytes()
			if !bytes.Equal(tt.output, encodedBytes) {
				t.Errorf("encode() = %v, want %v", tt.output, encodedBytes)
			}
		})
	}
}
