package ssz

import (
	"bytes"
	"testing"
)

type signingRootTest struct {
	Val1 interface{}
	Val2 interface{}
}

type truncateSignatureCase struct {
	Slot              uint64
	PreviousBlockRoot []byte
	Signature         []byte
}

type truncateLastCase struct {
	Slot           uint64
	StateRoot      []byte
	TruncatedField []byte
}

func TestSigningRoot(t *testing.T) {
	var signingRootTests = []signingRootTest{
		{
			Val1: &truncateSignatureCase{Slot: 20, Signature: []byte{'A', 'B'}},
			Val2: &truncateSignatureCase{Slot: 20, Signature: []byte("TESTING")},
		},
		{
			Val1: &truncateSignatureCase{
				Slot:              10,
				PreviousBlockRoot: []byte{'a', 'b'},
				Signature:         []byte("TESTINGDIFF")},
			Val2: &truncateSignatureCase{
				Slot:              10,
				PreviousBlockRoot: []byte{'a', 'b'},
				Signature:         []byte("TESTING23")},
		},
		{
			Val1: truncateSignatureCase{Slot: 50, Signature: []byte("THIS")},
			Val2: truncateSignatureCase{Slot: 50, Signature: []byte("DOESNT")},
		},
		{
			Val1: truncateSignatureCase{Signature: []byte("MATTER")},
			Val2: truncateSignatureCase{Signature: []byte("TESTING")},
		},
		{
			Val1: truncateLastCase{
				Slot:           5,
				StateRoot:      []byte("MATTERS"),
				TruncatedField: []byte("DOESNT MATTER"),
			},
			Val2: truncateLastCase{
				Slot:           5,
				StateRoot:      []byte("MATTERS"),
				TruncatedField: []byte("SHOULDNT MATTER"),
			},
		},
		{
			Val1: truncateLastCase{
				Slot:           550,
				StateRoot:      []byte("SHOULD"),
				TruncatedField: []byte("DOESNT"),
			},
			Val2: truncateLastCase{
				Slot:           550,
				StateRoot:      []byte("SHOULD"),
				TruncatedField: []byte("SHOULDNT"),
			},
		},
	}

	for i, test := range signingRootTests {
		output1, err := SigningRoot(test.Val1)
		if err != nil {
			t.Errorf("could not get the signing root of test %d, value 1 %v", i, err)
		}
		output2, err := SigningRoot(test.Val2)
		if err != nil {
			t.Errorf("could not get the signing root of test %d, value 2 %v", i, err)
		}
		// Check values have same result hash
		if !bytes.Equal(output1[:], output2[:]) {
			t.Errorf("test %d: hash mismatch: %X\n != %X", i, output1, output2)
		}
	}
}
