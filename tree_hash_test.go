package ssz

import (
	"bytes"
	"encoding/hex"
	"testing"

	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	"github.com/prysmaticlabs/prysm/shared/params"
)

func TestHashTreeRoot(t *testing.T) {
	fork := &fork{
		PreviousVersion: [4]byte{159, 65, 189, 91},
		CurrentVersion:  [4]byte{203, 176, 241, 215},
		Epoch:           11971467576204192310,
	}
	want, err := hex.DecodeString("3ad1264c33bc66b43a49b1258b88f34b8dbfa1649f17e6df550f589650d34992")
	if err != nil {
		t.Fatal(err)
	}
	root, err := HashTreeRoot(fork)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(root[:], want) {
		t.Errorf("want %v, HashTreeRoot() = %v", want, root)
	}
}

func BenchmarkHashTreeRoot_FullBlock(b *testing.B) {
	testConfig := params.BeaconConfig()
	testConfig.MaxTransfers = 1

	validatorCount := params.BeaconConfig().DepositsForChainStart * 4
	validators := make([]*pb.Validator, validatorCount)
	for i := 0; i < len(validators); i++ {
		validators[i] = &pb.Validator{
			EffectiveBalance:           params.BeaconConfig().MaxDepositAmount,
			ExitEpoch:                  params.BeaconConfig().FarFutureEpoch,
			WithdrawableEpoch:          params.BeaconConfig().FarFutureEpoch,
			ActivationEligibilityEpoch: params.BeaconConfig().FarFutureEpoch,
		}
	}
	validatorBalances := make([]uint64, len(validators))
	for i := 0; i < len(validatorBalances); i++ {
		validatorBalances[i] = params.BeaconConfig().MaxDepositAmount
	}

	randaoMixes := make([][]byte, params.BeaconConfig().LatestRandaoMixesLength)
	for i := 0; i < len(randaoMixes); i++ {
		randaoMixes[i] = params.BeaconConfig().ZeroHash[:]
	}

	var crosslinks []*pb.Crosslink
	for i := uint64(0); i < params.BeaconConfig().ShardCount; i++ {
		crosslinks = append(crosslinks, &pb.Crosslink{
			Epoch:    0,
			DataRoot: []byte{'A'},
		})
	}

	proposerSlashings := []*pb.ProposerSlashing{
		{
			ProposerIndex: 1,
			Header_1: &pb.BeaconBlockHeader{
				Slot:      0,
				Signature: []byte("A"),
			},
			Header_2: &pb.BeaconBlockHeader{
				Slot:      0,
				Signature: []byte("B"),
			},
		},
	}

	attesterSlashings := []*pb.AttesterSlashing{
		{
			Attestation_1: &pb.IndexedAttestation{
				Data: &pb.AttestationData{
					Crosslink: &pb.Crosslink{
						Shard: 5,
					},
				},
				CustodyBit_0Indices: []uint64{2, 3},
			},
			Attestation_2: &pb.IndexedAttestation{
				Data: &pb.AttestationData{
					Crosslink: &pb.Crosslink{
						Shard: 5,
					},
				},
				CustodyBit_0Indices: []uint64{2, 3},
			},
		},
	}
	// Set up transfer object for block
	transfers := []*pb.Transfer{
		{
			Slot:      1,
			Sender:    3,
			Recipient: 4,
			Fee:       params.BeaconConfig().MinDepositAmount,
			Amount:    params.BeaconConfig().MinDepositAmount,
			Pubkey:    []byte("A"),
		},
	}

	attestations := make([]*pb.Attestation, params.BeaconConfig().MaxAttestations)
	for i := 0; i < len(attestations); i++ {
		attestations[i] = &pb.Attestation{
			Data: &pb.AttestationData{
				SourceRoot: []byte("tron-sucks"),
				Crosslink: &pb.Crosslink{
					Shard:      uint64(i),
					ParentRoot: []byte("parent-root"),
					DataRoot:   params.BeaconConfig().ZeroHash[:],
				},
			},
			AggregationBitfield: []byte{0xC0, 0xC0, 0xC0, 0xC0, 0xC0, 0xC0, 0xC0,
				0xC0, 0xC0, 0xC0, 0xC0, 0xC0, 0xC0, 0xC0, 0xC0, 0xC0},
			CustodyBitfield: []byte{},
		}
	}

	blk := &pb.BeaconBlock{
		Slot: 1,
		Body: &pb.BeaconBlockBody{
			Eth1Data: &pb.Eth1Data{
				DepositRoot: []byte("trie-root"),
				BlockRoot:   []byte("trie-root"),
			},
			Attestations:      attestations,
			ProposerSlashings: proposerSlashings,
			AttesterSlashings: attesterSlashings,
			Transfers:         transfers,
		},
	}
	for n := 0; n < b.N; n++ {
		if _, err := HashTreeRoot(blk); err != nil {
			b.Fatal(err)
		}
	}
}
