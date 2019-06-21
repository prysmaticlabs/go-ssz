package autogenerated

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/prysmaticlabs/go-ssz"
)

func TestYaml(t *testing.T) {
	file, err := ioutil.ReadFile("ssz_minimal_one_formatted.yaml")
	if err != nil {
		t.Fatalf("Could not load file %v", err)
	}

	s := &SszMinimalTest{}
	if err := yaml.Unmarshal(file, s); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	for _, testCase := range s.TestCases {
		if !isEmpty(testCase.Attestation.Value) {
			encoded, err := ssz.Marshal(testCase.Attestation.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.Attestation.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.Attestation.Root) {
				t.Fatalf("Expected attestation %#x, received %#x", testCase.Attestation.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.Attestation.Serialized) {
				t.Fatalf("Expected attestation %#x, received %#x", testCase.Attestation.Serialized, encoded)
			}

		}
		if !isEmpty(testCase.AttestationData.Value) {
			encoded, err := ssz.Marshal(testCase.AttestationData.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.AttestationData.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.AttestationData.Root) {
				t.Fatalf("Expected attestation data %#x, received %#x", testCase.AttestationData.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.AttestationData.Serialized) {
				t.Fatalf("Expected attestation data %#x, received %#x", testCase.AttestationData.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.AttestationDataAndCustodyBit.Value) {
			encoded, err := ssz.Marshal(testCase.AttestationDataAndCustodyBit.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.AttestationDataAndCustodyBit.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.AttestationDataAndCustodyBit.Root) {
				t.Fatalf("Expected attestation custody bit %#x, received %#x", testCase.AttestationDataAndCustodyBit.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.AttestationDataAndCustodyBit.Serialized) {
				t.Fatalf("Expected attestation custody bit %#x, received %#x", testCase.AttestationDataAndCustodyBit.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.AttesterSlashing.Value) {
			encoded, err := ssz.Marshal(testCase.AttesterSlashing.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.AttesterSlashing.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.AttesterSlashing.Root) {
				t.Fatalf("Expected attester slashing %#x, received %#x", testCase.AttesterSlashing.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.AttesterSlashing.Serialized) {
				t.Fatalf("Expected attester slashing %#x, received %#x", testCase.AttesterSlashing.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.BeaconBlock.Value) {
			encoded, err := ssz.Marshal(testCase.BeaconBlock.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.BeaconBlock.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.BeaconBlock.Root) {
				t.Fatalf("Expected beacon block %#x, received %#x", testCase.BeaconBlock.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.BeaconBlock.Serialized) {
				t.Fatalf("Expected beacon block %#x, received %#x", testCase.BeaconBlock.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.BeaconBlockBody.Value) {
			encoded, err := ssz.Marshal(testCase.BeaconBlockBody.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.BeaconBlockBody.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.BeaconBlockBody.Root) {
				t.Fatalf("Expected %#x, received %#x", testCase.BeaconBlockBody.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.BeaconBlockBody.Serialized) {
				t.Fatalf("Expected %#x, received %#x", testCase.BeaconBlockBody.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.BeaconBlockHeader.Value) {
			encoded, err := ssz.Marshal(testCase.BeaconBlockHeader.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.BeaconBlockHeader.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.BeaconBlockHeader.Root) {
				t.Fatalf("Expected block header %#x, received %#x", testCase.BeaconBlockHeader.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.BeaconBlockHeader.Serialized) {
				t.Fatalf("Expected block header %#x, received %#x", testCase.BeaconBlockHeader.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.BeaconState.Value) {
			encoded, err := ssz.Marshal(testCase.BeaconState.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.BeaconState.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.BeaconState.Root) {
				t.Fatalf("Expected beacon state %#x, received %#x", testCase.BeaconState.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.BeaconState.Serialized) {
               t.Fatal("Serializations do not match")
			}
		}
		if !isEmpty(testCase.Crosslink.Value) {
			encoded, err := ssz.Marshal(testCase.Crosslink.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.Crosslink.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.Crosslink.Root) {
				t.Fatalf("Expected crosslink %#x, received %#x", testCase.Crosslink.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.Crosslink.Serialized) {
				t.Fatalf("Expected crosslink %#x, received %#x", testCase.Crosslink.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.Deposit.Value) {
			encoded, err := ssz.Marshal(testCase.Deposit.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.Deposit.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.Deposit.Root) {
				t.Fatalf("Expected deposit %#x, received %#x", testCase.Deposit.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.Deposit.Serialized) {
				t.Fatalf("Expected deposit %#x, received %#x", testCase.Deposit.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.DepositData.Value) {
			encoded, err := ssz.Marshal(testCase.DepositData.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.DepositData.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.DepositData.Root) {
				t.Fatalf("Expected deposit data %#x, received %#x", testCase.DepositData.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.DepositData.Serialized) {
				t.Fatalf("Expected deposit data %#x, received %#x", testCase.DepositData.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.Eth1Data.Value) {
			encoded, err := ssz.Marshal(testCase.Eth1Data.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.Eth1Data.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.Eth1Data.Root) {
				t.Fatalf("Expected eth1data %#x, received %#x", testCase.Eth1Data.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.Eth1Data.Serialized) {
				t.Fatalf("Expected eth1data %#x, received %#x", testCase.Eth1Data.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.Fork.Value) {
			encoded, err := ssz.Marshal(testCase.Fork.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.Fork.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.Fork.Root) {
				t.Errorf("Expected fork %#x, received %#x", testCase.Fork.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.Fork.Serialized) {
				t.Errorf("Expected fork %v, received %v", testCase.Fork.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.HistoricalBatch.Value) {
			encoded, err := ssz.Marshal(testCase.HistoricalBatch.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.HistoricalBatch.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.HistoricalBatch.Root) {
				t.Fatalf("Expected historical batch %#x, received %#x", testCase.HistoricalBatch.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.HistoricalBatch.Serialized) {
				t.Fatalf("Expected historical batch %#x, received %#x", testCase.HistoricalBatch.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.IndexedAttestation.Value) {
			encoded, err := ssz.Marshal(testCase.IndexedAttestation.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.IndexedAttestation.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.IndexedAttestation.Root) {
				t.Fatalf("Expected indexed att %#x, received %#x", testCase.IndexedAttestation.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.IndexedAttestation.Serialized) {
				t.Fatalf("Expected indexed att %#x, received %#x", testCase.IndexedAttestation.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.PendingAttestation.Value) {
			encoded, err := ssz.Marshal(testCase.PendingAttestation.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.PendingAttestation.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.PendingAttestation.Root) {
				t.Fatalf("Expected pending att %#x, received %#x", testCase.PendingAttestation.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.PendingAttestation.Serialized) {
				t.Fatalf("Expected pending att %#x, received %#x", testCase.PendingAttestation.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.ProposerSlashing.Value) {
			encoded, err := ssz.Marshal(testCase.ProposerSlashing.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.ProposerSlashing.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.ProposerSlashing.Root) {
				t.Fatalf("Expected proposer slashing %#x, received %#x", testCase.ProposerSlashing.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.ProposerSlashing.Serialized) {
				t.Fatalf("Expected proposer slashing %#x, received %#x", testCase.ProposerSlashing.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.Transfer.Value) {
			encoded, err := ssz.Marshal(testCase.Transfer.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.Transfer.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.Transfer.Root) {
				t.Fatalf("Expected transfer %#x, received %#x", testCase.Transfer.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.Transfer.Serialized) {
				t.Fatalf("Expected transfer %#x, received %#x", testCase.Transfer.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.Validator.Value) {
			encoded, err := ssz.Marshal(testCase.Validator.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.Validator.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.Validator.Root) {
				t.Fatalf("Expected validator %#x, received %#x", testCase.Validator.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.Validator.Serialized) {
				t.Fatalf("Expected validator %#x, received %#x", testCase.Validator.Serialized, encoded)
			}
		}
		if !isEmpty(testCase.VoluntaryExit.Value) {
			encoded, err := ssz.Marshal(testCase.VoluntaryExit.Value)
			if err != nil {
				t.Fatal(err)
			}
			root, err := ssz.HashTreeRoot(testCase.VoluntaryExit.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(root[:], testCase.VoluntaryExit.Root) {
				t.Fatalf("Expected voluntary exit %#x, received %#x", testCase.VoluntaryExit.Root, root[:])
			}
			if !bytes.Equal(encoded, testCase.VoluntaryExit.Serialized) {
				t.Fatalf("Expected voluntary exit %#x, received %#x", testCase.VoluntaryExit.Serialized, encoded)
			}
		}
	}
}

func isEmpty(item interface{}) bool {
	val := reflect.ValueOf(item)
	for i := 0; i < val.NumField(); i++ {
		if !reflect.DeepEqual(val.Field(i).Interface(), reflect.Zero(val.Field(i).Type()).Interface()) {
			return false
		}
	}
	return true
}