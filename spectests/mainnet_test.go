package spectests

import (
	"bytes"
	"io/ioutil"
	"path"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/prysmaticlabs/go-ssz"
)

func TestSSZStatic_Mainnet(t *testing.T) {
	testFolders, testsFolderPath := testFolders(t, "mainnet", "ssz_static")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath, err := bazel.Runfile(path.Join(testsFolderPath, folder.Name()))
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}
			innerSSZFolders, err := ioutil.ReadDir(folderPath)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}
			for _, innerFolder := range innerSSZFolders {
				innerPath := path.Join(folderPath, innerFolder.Name())
				cases, err := ioutil.ReadDir(innerPath)
				if err != nil {
					t.Fatalf("Failed to read file: %v", err)
				}
				for _, cs := range cases {
					serialized, err := bazelFileBytes(path.Join(innerPath, cs.Name(), "serialized.ssz"))
					if err != nil {
						t.Fatal(err)
					}
					switch folder.Name() {
					case "AggregateAndProof":
						dec := &MainnetAggregateAndProof{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "Attestation":
						dec := &MainnetAttestation{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "AttestationData":
						dec := &MainnetAttestationData{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "AttestationDataAndCustodyBit":
						dec := &MainnetAttestationAndCustodyBit{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "AttesterSlashing":
						dec := &MainnetAttesterSlashing{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "BeaconBlock":
						dec := &MainnetBlock{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "BeaconBlockBody":
						dec := &MainnetBlockBody{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "BeaconBlockHeader":
						dec := &MainnetBlockHeader{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "BeaconState":
						dec := &MainnetBeaconState{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "Checkpoint":
						dec := &MainnetCheckpoint{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "Deposit":
						dec := &MainnetDeposit{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "DepositData":
						dec := &MainnetDepositData{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "Eth1Data":
						dec := &MainnetEth1Data{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "Fork":
						dec := &MainnetFork{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "HistoricalBatch":
						dec := &MainnetHistoricalBatch{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "IndexedAttestation":
						dec := &MainnetIndexedAttestation{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "PendingAttestation":
						dec := &MainnetPendingAttestation{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "ProposerSlashing":
						dec := &MainnetProposerSlashing{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "Validator":
						dec := &MainnetValidator{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					case "VoluntaryExit":
						dec := &MainnetVoluntaryExit{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, enc) {
							t.Errorf("Wanted %v, received %v", serialized, enc)
						}
					}
				}
			}
		})
	}
}
