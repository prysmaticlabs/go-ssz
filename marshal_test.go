package ssz

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"
)

type testDepositData struct {
	Pubkey                []byte   `protobuf:"bytes,1,opt,name=pubkey,proto3" json:"pubkey,omitempty" ssz-size:"48"`
	WithdrawalCredentials []byte   `protobuf:"bytes,2,opt,name=withdrawal_credentials,json=withdrawalCredentials,proto3" json:"withdrawal_credentials,omitempty" ssz-size:"32"`
	Amount                uint64   `protobuf:"varint,3,opt,name=amount,proto3" json:"amount,omitempty"`
	Signature             []byte   `protobuf:"bytes,4,opt,name=signature,proto3" json:"signature,omitempty" ssz-size:"96"`
	XXXNoUnkeyedLiteral   struct{} `json:"-"`
	XXXunrecognized       []byte   `json:"-"`
	XXXsizecache          int32    `json:"-"`
}

func TestMarshalDepositDataLen(t *testing.T) {
	expectedResult, err := hex.DecodeString("b334e2fce27e6ebeaba7bbb9ca112c7f6acf5e863b22db3eec99458ddb1d5ac1a05f177590ae5257385a3bb7573405c60055dce78f09d5867624ec74bb486e64e15e6335537e926917fffd7c4d8f94900040597307000000982f0411aa3802331ecdcee8e1e39d14cf53d878bc4ca713b4d21311454f2a9c277bf6a27daa63c01c1e017eb355ed2f1928cd936ae37feee6cd8b339c92739d607fa77dea7bb98092f2c61c6d5df0aa4850614801eece31befd9f217235d353")
	if err != nil {
		t.Fatalf("Cannot read expected result: %v", err)
	}
	testStruct := &testDepositData{
		Pubkey:                expectedResult[0:48],
		WithdrawalCredentials: expectedResult[48 : 48+32],
		Amount:                binary.LittleEndian.Uint64(expectedResult[48+32 : 48+32+8]),
		Signature:             expectedResult[48+32+8 : 48+32+8+96],
	}

	serializedData, err := Marshal(testStruct)
	if err != nil {
		t.Fatalf("Cannot marshal data: %v", err)
	}
	if len(serializedData) != 184 {
		t.Fatalf("invalid result len: %v", len(serializedData))
	}
	if !bytes.Equal(serializedData, expectedResult) {
		t.Fatalf(`invalid marshal result
			result  : %#x
			expected: %#x
		`, serializedData, expectedResult)
	}
}
