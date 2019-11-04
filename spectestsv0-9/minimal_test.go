package spectestsv0_9

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
)

func TestSSZStatic_Minimal(t *testing.T) {
	testsFolderPath := path.Join("tests", "minimal", "phase0", "ssz_static")
	newPath, err := bazel.Runfile(testsFolderPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(newPath)
	testFolders, testsFolderPath := testFolders(t, "minimal", "ssz_static")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			t.Log(folderPath)
		})
	}
}

func testFolders(t *testing.T, config string, folderPath string) ([]os.FileInfo, string) {
	testsFolderPath := path.Join("tests", config, "phase0", folderPath)
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
