package spectestsv0_9

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"

	"github.com/prysmaticlabs/go-ssz"
)

func TestSSZStatic_Minimal(t *testing.T) {
	testFolders, testsFolderPath := testFolders(t, "minimal", "ssz_static")
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
						dec := &MinimalAggregateAndProof{}
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
						dec := &MinimalAttestation{}
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
						dec := &MinimalAttestationData{}
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
						dec := &MinimalAttestationAndCustodyBit{}
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
						dec := &MinimalAttesterSlashing{}
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
						dec := &MinimalBlock{}
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
						dec := &MinimalBlockBody{}
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
						dec := &MinimalBlockHeader{}
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
						dec := &MinimalBeaconState{}
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
						dec := &MinimalCheckpoint{}
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
						dec := &MinimalDeposit{}
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
						dec := &MinimalDepositData{}
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
						dec := &MinimalEth1Data{}
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
						dec := &MinimalFork{}
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
						dec := &MinimalHistoricalBatch{}
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
						dec := &MinimalIndexedAttestation{}
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
						dec := &MinimalPendingAttestation{}
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
						dec := &MinimalProposerSlashing{}
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
						dec := &MinimalValidator{}
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
						dec := &MinimalVoluntaryExit{}
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

func testFolders(t *testing.T, config string, folderPath string) ([]os.FileInfo, string) {
	testsFolderPath := path.Join("eth2_spec_tests_minimal", "tests", config, "phase0", folderPath)
	filepath, err := bazel.Runfile(testsFolderPath)
	if err != nil {
		t.Fatal(err)
	}
	testFolders, err := ioutil.ReadDir(filepath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	return testFolders, testsFolderPath
}

func bazelFileBytes(filePaths ...string) ([]byte, error) {
	filepath, err := bazel.Runfile(path.Join(filePaths...))
	if err != nil {
		return nil, err
	}
	fileBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}
