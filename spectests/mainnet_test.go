package spectests

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/prysmaticlabs/go-ssz"
	experiment "github.com/prysmaticlabs/go-ssz/experiment"
	pb "github.com/prysmaticlabs/go-ssz/experiment/beacon/p2p/v1"
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
					rootPath := path.Join(innerPath, cs.Name(), "roots.yaml")
					cont := &rootContainer{}
					populateStructFromYaml(t, rootPath, cont)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
						}
					case "BeaconState":
						dec := &pb.BeaconState{}
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
						rt := experiment.StateRoot(dec)
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Fatalf("Wanted root %#x, got %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
						rt, err := ssz.HashTreeRoot(dec)
						if err != nil {
							t.Fatal(err)
						}
						if fmt.Sprintf("%#x", rt) != cont.Root {
							t.Errorf("Wanted %#x, received %#x", cont.Root, rt)
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
