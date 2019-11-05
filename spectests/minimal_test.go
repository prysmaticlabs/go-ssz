package spectests

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"

	"github.com/prysmaticlabs/go-ssz"
)

type rootContainer struct {
	Root string
}

func TestSSZStatic_Minimal(t *testing.T) {
	testFolders, testsFolderPath := testFolders(t, "minimal", "ssz_static")
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
				cases, err := ioutil.ReadDir(innerPath)
				if err != nil {
					t.Fatalf("Failed to read file: %v", err)
				}
				for _, cs := range cases {
					serializedPath := path.Join(innerPath, cs.Name(), "serialized.ssz")
					serialized, err := bazelFileBytes(serializedPath)
					if err != nil {
						t.Fatal(err)
					}
					rootPath := path.Join(innerPath, cs.Name(), "roots.yaml")
					cont := &rootContainer{}
					populateStructFromYaml(t, rootPath, cont)
					switch folder.Name() {
					case "AggregateAndProof":
						dec := &minimalAggregateAndProof{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "Attestation":
						dec := &minimalAttestation{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "AttestationData":
						dec := &minimalAttestationData{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "AttestationDataAndCustodyBit":
						dec := &minimalAttestationAndCustodyBit{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "AttesterSlashing":
						dec := &minimalAttesterSlashing{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "BeaconBlock":
						dec := &minimalBlock{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "BeaconBlockBody":
						dec := &minimalBlockBody{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "BeaconBlockHeader":
						dec := &minimalBlockHeader{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "BeaconState":
						dec := &minimalBeaconState{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "Checkpoint":
						dec := &minimalCheckpoint{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "Deposit":
						dec := &minimalDeposit{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "DepositData":
						dec := &minimalDepositData{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "Eth1Data":
						dec := &minimalEth1Data{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "Fork":
						dec := &minimalFork{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "HistoricalBatch":
						dec := &minimalHistoricalBatch{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "IndexedAttestation":
						dec := &minimalIndexedAttestation{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "PendingAttestation":
						dec := &minimalPendingAttestation{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "ProposerSlashing":
						dec := &minimalProposerSlashing{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "Validator":
						dec := &minimalValidator{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "VoluntaryExit":
						dec := &minimalVoluntaryExit{}
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					default:
						t.Error("Case not covered")
					}
				}
			}
		})
	}
}

func testFolders(t *testing.T, config string, folderPath string) ([]os.FileInfo, string) {
	testsFolderPath := path.Join("eth2_spec_tests_"+config, "tests", config, "phase0", folderPath)
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
