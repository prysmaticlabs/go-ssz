package ssz

import "testing"

type AmbiguousItem struct {
	Field1 []byte
	Field2 uint64
}

func TestDeepEqual(t *testing.T) {
	original := AmbiguousItem{
		Field2: 5,
	}
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded AmbiguousItem
	if err := Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if !DeepEqual(original, decoded) {
		t.Errorf("Expected %v, received %v", original, decoded)
	}
}
