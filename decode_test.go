package ssz

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDecodeUint(t *testing.T) {
	tests := []struct {
		input  interface{}
	}{
		{ input: uint64(5) },
		{ input: uint64(23929309) },
	}
	for _, tt := range tests {
		buffer := new(bytes.Buffer)
		if err := Encode(buffer, tt.input); err != nil {
			panic(err)
		}
		var decoded uint64
		if err := Decode(bytes.NewReader(buffer.Bytes()), &decoded); err != nil {
			t.Fatal(err)
		}
		fmt.Println(decoded)
		if decoded != tt.input {
			t.Errorf("Expected %d, received %d", tt.input, decoded)
		}
	}
}
