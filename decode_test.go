package ssz

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		input  interface{}
		ptr interface{}
	}{
		// Bool test cases.
		{ input: true, ptr: new(bool) },
		{ input: false, ptr: new(bool) },
		// Uint8 test cases.
		{ input: byte(1), ptr: new(byte) },
		{ input: byte(0), ptr: new(byte) },
		// Uint16 test cases.
		{ input: uint16(100), ptr: new(uint16) },
		{ input: uint16(232), ptr: new(uint16) },
		// Uint32 test cases.
		{ input: uint32(1), ptr: new(uint32) },
		{ input: uint32(1029391), ptr: new(uint32) },
		// Uint64 test cases.
		{ input: uint64(5), ptr: new(uint64) },
		{ input: uint64(23929309), ptr: new(uint64) },
	}
	for _, tt := range tests {
		buffer := new(bytes.Buffer)
		if err := Encode(buffer, tt.input); err != nil {
			panic(err)
		}
		fmt.Printf("Encoded: %v\n", buffer.Bytes())
		if err := Decode(bytes.NewReader(buffer.Bytes()), tt.ptr); err != nil {
			t.Fatal(err)
		}
		output := reflect.ValueOf(tt.ptr).Elem().Interface()
		fmt.Printf("Decoded: %v\n", output)
		if !reflect.DeepEqual(output, tt.input) {
			t.Errorf("Expected %d, received %d", tt.input, output)
		}
	}
}
