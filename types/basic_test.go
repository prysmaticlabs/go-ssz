package types

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/prysmaticlabs/go-bitfield"
)

type depData struct {
	Pubkey     []byte `ssz-size:"48"`
	Withdrawal []byte `ssz-size:"32"`
	Amount     uint64
	Signature  []byte `ssz-size:"96"`
}

type deposit struct {
	Proof [][]byte `ssz-size:"33,32"`
	Data  depData
}

type attestation struct {
	AggregationBits bitfield.Bitlist `ssz-max:"4096"`
	Data            attData
	CustodyBits     bitfield.Bitlist `ssz-max:"4096"`
	Signature       []byte           `ssz-size:"96"`
}

type attData struct {
	Root      []byte `ssz-size:"32"`
	Source    checkpoint
	Target    checkpoint
	Crosslink crosslink
}

type checkpoint struct {
	Epoch uint64
	Root  []byte `ssz-size:"32"`
}

type crosslink struct {
	Shard      uint64
	Parent     []byte `ssz-size:"32"`
	StartEpoch uint64
	EndEpoch   uint64
	DataRoot   []byte `ssz-size:"32"`
}

func TestBasicSSZ_Root(t *testing.T) {
	//roots := make([][]byte, 33)
	emptyRoot := make([]byte, 32)
	//emptyKey := make([]byte, 48)
	emptySig := make([]byte, 96)
	//for i := 0; i < 32; i++ {
	//	roots[i] = emptyRoot
	//}
	//item := deposit{
	//	Proof: roots,
	//	Data: depData{
	//		Pubkey:     emptyKey,
	//		Withdrawal: emptyRoot,
	//		Amount:     0,
	//		Signature:  emptySig,
	//	},
	//}
	item := attestation{
		AggregationBits: []byte{1},
		Data: attData{
			Root: emptyRoot,
			Source: checkpoint{
				Epoch: 0,
				Root:  emptyRoot,
			},
			Target: checkpoint{
				Epoch: 0,
				Root:  emptyRoot,
			},
			Crosslink: crosslink{
				Shard:      0,
				Parent:     emptyRoot,
				StartEpoch: 0,
				EndEpoch:   0,
				DataRoot:   emptyRoot,
			},
		},
		CustodyBits: []byte{2},
		Signature:   emptySig,
	}
	val := reflect.ValueOf(item)
	typ := reflect.TypeOf(item)
	factory, _ := SSZFactory(val, typ)
	res, err := factory.Root(val, typ, 0)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#x\n", res)
}
