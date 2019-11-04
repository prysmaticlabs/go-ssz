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
					case "BeaconBlockHeader":
						dec := &MinimalBlockHeader{}
						if err := ssz.Unmarshal(serialized, dec); err != nil {
							t.Fatal(err)
						}
						enc, err := ssz.Marshal(dec)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(serialized, serialized) {
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
						if !bytes.Equal(serialized, serialized) {
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
						if !bytes.Equal(serialized, serialized) {
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
