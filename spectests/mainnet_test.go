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
		eth2TypeName := folder.Name()
		t.Run(eth2TypeName, func(t *testing.T) {
			sszCases, err := bazel.Runfile(path.Join(testsFolderPath, eth2TypeName))
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}
			innerSSZFolders, err := ioutil.ReadDir(sszCases)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}
			for _, innerFolder := range innerSSZFolders {
				innerPath := path.Join(sszCases, innerFolder.Name())
				testCases, err := ioutil.ReadDir(innerPath)
				if err != nil {
					t.Fatalf("Failed to read file: %v", err)
				}
				for _, cs := range testCases {
					serialized, err := bazelFileBytes(path.Join(innerPath, cs.Name(), "serialized.ssz"))
					if err != nil {
						t.Fatal(err)
					}
					switch folder.Name() {
					case "AggregateAndProof":
						dec := &mainnetAggregateAndProof{}
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
						dec := &mainnetAttestation{}
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
						dec := &mainnetAttestationData{}
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
						dec := &mainnetAttestationAndCustodyBit{}
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
						dec := &mainnetAttesterSlashing{}
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
						dec := &mainnetBlock{}
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
						dec := &mainnetBlockBody{}
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
						dec := &mainnetBlockHeader{}
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
						dec := &mainnetBeaconState{}
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
						dec := &mainnetCheckpoint{}
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
						dec := &mainnetDeposit{}
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
						dec := &mainnetDepositData{}
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
						dec := &mainnetEth1Data{}
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
						dec := &mainnetFork{}
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
						dec := &mainnetHistoricalBatch{}
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
						dec := &mainnetIndexedAttestation{}
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
						dec := &mainnetPendingAttestation{}
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
						dec := &mainnetProposerSlashing{}
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
						dec := &mainnetValidator{}
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
						dec := &mainnetVoluntaryExit{}
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
