package ssz

import (
	"bytes"
	"testing"
)

type signingRootTest struct {
	val1 interface{}
	val2 interface{}
}

type truncateSignatureCase struct {
	slot              uint64
	previousBlockRoot []byte
	signature         []byte
}

type truncateLastCase struct {
	slot           uint64
	stateRoot      []byte
	truncatedField []byte
}

func TestSigningRoot(t *testing.T) {
	var signingRootTests = []signingRootTest{
		{
			val1: &truncateSignatureCase{slot: 20, signature: []byte{'A', 'B'}},
			val2: &truncateSignatureCase{slot: 20, signature: []byte("TESTING")},
		},
		{
			val1: &truncateSignatureCase{
				slot:              10,
				previousBlockRoot: []byte{'a', 'b'},
				signature:         []byte("TESTINGDIFF")},
			val2: &truncateSignatureCase{
				slot:              10,
				previousBlockRoot: []byte{'a', 'b'},
				signature:         []byte("TESTING23")},
		},
		{
			val1: truncateSignatureCase{slot: 50, signature: []byte("THIS")},
			val2: truncateSignatureCase{slot: 50, signature: []byte("DOESNT")},
		},
		{
			val1: truncateSignatureCase{signature: []byte("MATTER")},
			val2: truncateSignatureCase{signature: []byte("TESTING")},
		},
		{
			val1: truncateLastCase{
				slot:           5,
				stateRoot:      []byte("MATTERS"),
				truncatedField: []byte("DOESNT MATTER"),
			},
			val2: truncateLastCase{
				slot:           5,
				stateRoot:      []byte("MATTERS"),
				truncatedField: []byte("SHOULDNT MATTER"),
			},
		},
		{
			val1: truncateLastCase{
				slot:           550,
				stateRoot:      []byte("SHOULD"),
				truncatedField: []byte("DOESNT"),
			},
			val2: truncateLastCase{
				slot:           550,
				stateRoot:      []byte("SHOULD"),
				truncatedField: []byte("SHOULDNT"),
			},
		},
	}

	for i, test := range signingRootTests {
		output1, err := SigningRoot(test.val1)
		if err != nil {
			t.Fatalf("could not get the signing root of test %d, value 1 %v", i, err)
		}
		output2, err := SigningRoot(test.val2)
		if err != nil {
			t.Fatalf("could not get the signing root of test %d, value 2 %v", i, err)
		}
		// Check values have same result hash
		if !bytes.Equal(output1[:], output2[:]) {
			t.Errorf("test %d: hash mismatch: %X\n != %X", i, output1, output2)
		}
	}
}
