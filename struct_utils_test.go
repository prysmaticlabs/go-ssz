package ssz

import (
	"reflect"
	"testing"
)

type structWithTags struct {
	NestedItem    [][][][]byte `ssz:"size=4,1,4,1"`
	UnboundedItem [][]byte     `ssz:"size=?,4"`
}

func TestInferTypeFromStructTags(t *testing.T) {
	structExample := structWithTags{
		NestedItem:    [][][][]byte{{{{4}}}, {{{3}}}},
		UnboundedItem: [][]byte{{2, 3, 4, 5}, {1, 2, 3, 4}},
	}
	// We then verify that highly nested items can have their types
	// inferred via SSZ field tags.
	typ := reflect.TypeOf(structExample)
	sizes, exists, err := parseSSZFieldTags(typ.Field(0))
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("Expected struct tags to exist")
	}
	fType := inferFieldTypeFromSizeTags(typ.Field(0), sizes)
	expectedField := [4][1][4][1]byte{}
	expectedFieldType := reflect.TypeOf(expectedField)

	if expectedFieldType != fType {
		t.Errorf("Expected inferred field: %v, received %v", expectedFieldType, fType)
	}

	// We then verify that unbounded items can be formed via SSZ field tags
	// and their type can be correctly inferred.
	sizes, exists, err = parseSSZFieldTags(typ.Field(1))
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("Expected struct tags to exist")
	}
	fType = inferFieldTypeFromSizeTags(typ.Field(1), sizes)
	unboundedExpectedField := [][4]byte{}
	unboundedExpectedFieldType := reflect.TypeOf(unboundedExpectedField)

	if unboundedExpectedFieldType != fType {
		t.Errorf("Expected inferred field: %v, received %v", expectedFieldType, fType)
	}
}
