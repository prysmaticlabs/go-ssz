package spectests

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/ghodss/yaml"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/prysmaticlabs/go-ssz"
)

func TestSSZGenericSpecTests(t *testing.T) {
	fullPath := "eth2_spec_tests_general/tests/general/phase0/ssz_generic/"
	filepath, err := bazel.Runfile(fullPath)
	if err != nil {
		t.Fatal(err)
	}
	testFolders, err := ioutil.ReadDir(filepath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	for _, folder := range testFolders {
		// Runs the valid SSZ tests
		runSSZGenericTests(t, filepath, folder.Name(), "valid")
		// Runs the invalid SSZ tests
		runSSZGenericTests(t, filepath, folder.Name(), "invalid")
	}
}

func runSSZGenericTests(t *testing.T, testPath string, dataType string, validityFolder string) {
	valid := validityFolder == "valid"

	filepath, err := bazel.Runfile(path.Join(testPath, dataType, validityFolder))
	if err != nil {
		t.Fatal(err)
	}
	testFolders, err := ioutil.ReadDir(filepath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	for _, testFolder := range testFolders {
		tFName := testFolder.Name()
		// testLogName := fmt.Sprintf("%s/%s_%d", dataType, validityFolder, i)
		serializedBytes, err := FileBytesOrDie(path.Join(filepath, tFName, "serialized.ssz"))
		if err != nil {
			t.Fatal(err)
		}

		yamlPath := path.Join(filepath, tFName, "meta.yaml")

		fmt.Println(tFName)

		switch dataType {
		case "basic_vector":
			runBasicVectorSSZTestCase(t, serializedBytes, yamlPath, tFName, valid)
		case "bitlist":
			sizeStr := strings.Split(tFName, "_")[1]
			var size int
			if sizeStr != "no" {
				size, err = strconv.Atoi(sizeStr)
				if err != nil {
					t.Fatalf("could not get convert string to int: %v", err)
				}
			} else {
				size = 0
			}
			runBitlistSSZTestCase(t, serializedBytes, uint64(size), yamlPath, valid)
		case "bitvector":
			runBitvectorSSZTestCase(t, serializedBytes, yamlPath, tFName, valid)
		case "boolean":
			runBooleanSSZTestCase(t, serializedBytes, yamlPath, valid)
		// case "containers":
		// 	truncatedName := tFName[:strings.Index(tFName, "_")]
		// 	runContainerSSZTestCase(t, serializedBytes, yamlPath, truncatedName, valid)
		case "uints":
			truncatedName := tFName[:strings.Index(tFName[4:], "_")]
			runUintSSZTestCase(t, serializedBytes, yamlPath, truncatedName, valid)
		}
	}
}

func runBasicVectorSSZTestCase(t *testing.T, objBytes []byte, yamlPath string, testName string, valid bool) {
	testInfo := strings.Split(testName, "_")[1:]
	// if !valid && len(testInfo) > 4 && testInfo[5] == "more" {
	// 	t.Fatalf("%#x", objBytes)
	// }
	switch testInfo[0] {
	case "bool":
		switch testInfo[1] {
		case "0":
			var result [0]bool
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform bool array ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform bool array root check: %v", err)
			}
		case "1":
			var result [1]bool
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform bool array ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform bool array root check: %v", err)
			}
		case "2":
			var result [2]bool
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform bool array ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform bool array root check: %v", err)
			}
		case "3":
			var result [3]bool
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform bool array ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform bool array root check: %v", err)
			}
		case "4":
			var result [4]bool
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform bool array ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform bool array root check: %v", err)
			}
		case "5":
			var result [5]bool
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform bool array ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform bool array root check: %v", err)
			}
		case "8":
			var result [8]bool
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform bool array ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform bool array root check: %v", err)
			}
		case "16":
			var result [16]bool
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform bool array ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform bool array root check: %v", err)
			}
		case "31":
			var result [31]bool
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform bool array ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform bool array root check: %v", err)
			}
		case "512":
			var result [512]bool
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform bool array ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform bool array root check: %v", err)
			}
		case "513":
			var result [513]bool
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform bool array ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform bool array root check: %v", err)
			}
		}
	case "uint8":
		switch testInfo[1] {
		case "0":
			var result [0]uint8
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "1":
			var result [1]uint8
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "2":
			var result [2]uint8
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "3":
			var result [3]uint8
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "4":
			var result [4]uint8
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "5":
			var result [5]uint8
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "8":
			var result [8]uint8
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "16":
			var result [16]uint8
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "31":
			var result [31]uint8
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "512":
			var result [512]uint8
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "513":
			var result [513]uint8
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		}
	case "uint16":
		switch testInfo[1] {
		case "0":
			var result [0]uint16
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "1":
			var result [1]uint16
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "2":
			var result [2]uint16
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "3":
			var result [3]uint16
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "4":
			var result [4]uint16
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "5":
			var result [5]uint16
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "8":
			var result [8]uint16
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "16":
			var result [16]uint16
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "31":
			var result [31]uint16
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "512":
			var result [512]uint16
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "513":
			var result [513]uint16
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		}
	case "uint32":
		switch testInfo[1] {
		case "0":
			var result [0]uint32
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "1":
			var result [1]uint32
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "2":
			var result [2]uint32
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "3":
			var result [3]uint32
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "4":
			var result [4]uint32
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "5":
			var result [5]uint32
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "8":
			var result [8]uint32
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "16":
			var result [16]uint32
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "31":
			var result [31]uint32
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "512":
			var result [512]uint32
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "513":
			var result [513]uint32
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		}
	case "uint64":
		switch testInfo[1] {
		case "0":
			var result [0]uint64
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "1":
			var result [1]uint64
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "2":
			var result [2]uint64
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "3":
			var result [3]uint64
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "4":
			var result [4]uint64
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "5":
			var result [5]uint64
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "8":
			var result [8]uint64
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "16":
			var result [16]uint64
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "31":
			var result [31]uint64
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "512":
			var result [512]uint64
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		case "513":
			var result [513]uint64
			if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
				t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
			}
			if err := PerformRootCheck(result, yamlPath, valid); err != nil {
				t.Fatalf("Could not perform root check for case %s: %v", testName, err)
			}
		}
	case "uint128":
		return
	case "uint256":
		return
	}
}

func runBitlistSSZTestCase(t *testing.T, objBytes []byte, size uint64, yamlPath string, valid bool) {
	if !valid {
		return
	}
	var result bitfield.Bitlist
	if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
		t.Fatalf("Could not perform bitlist ssz test case: %v", err)
	}
	if err := PerformBitlistRootCheck(result, size, yamlPath, valid); err != nil {
		t.Fatalf("Could not perform bitlist root check: %v", err)
	}
}

func runBitvectorSSZTestCase(t *testing.T, objBytes []byte, yamlPath string, testName string, valid bool) {
	if !valid {
		return
	}
	testInfo := strings.Split(testName, "_")[1:]
	if testInfo[0] == "4" {
		var result bitfield.Bitvector4
		if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
			t.Fatalf("Could not perform bitvector ssz test case: %v", err)
		}
		// if err := PerformRootCheck(result, yamlPath, valid); err != nil {
		// 	t.Fatalf("Could not perform bitvector root check: %v", err)
		// }
	}
}

func runBooleanSSZTestCase(t *testing.T, objBytes []byte, yamlPath string, valid bool) {
	var result bool
	if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
		t.Fatalf("Could not perform bool ssz test case: %v", err)
	}
	if err := PerformRootCheck(result, yamlPath, valid); err != nil {
		t.Fatalf("Could not perform bool root check: %v", err)
	}
}

func runContainerSSZTestCase(t *testing.T, objBytes []byte, yamlPath string, testName string, valid bool) {
	switch testName {
	case "ComplexTestStruct":
		var container complexTestStruct
		if err := PerformSSZCheck(objBytes, &container, yamlPath, valid); err != nil {
			t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
		}
		if err := PerformRootCheck(container, yamlPath, valid); err != nil {
			t.Fatalf("Could not perform root check for case %s: %v", testName, err)
		}
	case "FixedTestStruct":
		var container fixedTestStruct
		if err := PerformSSZCheck(objBytes, &container, yamlPath, valid); err != nil {
			t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
		}
		if err := PerformRootCheck(container, yamlPath, valid); err != nil {
			t.Fatalf("Could not perform root check for case %s: %v", testName, err)
		}
	case "SingleFieldTestStruct":
		var container singleFieldStruct
		if err := PerformSSZCheck(objBytes, &container, yamlPath, valid); err != nil {
			t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
		}
		if err := PerformRootCheck(container, yamlPath, valid); err != nil {
			t.Fatalf("Could not perform root check for case %s: %v", testName, err)
		}
	case "SmallTestStruct":
		var container smallTestStruct
		if err := PerformSSZCheck(objBytes, &container, yamlPath, valid); err != nil {
			t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
		}
		if err := PerformRootCheck(container, yamlPath, valid); err != nil {
			t.Fatalf("Could not perform root check for case %s: %v", testName, err)
		}
	case "VarTestStruct":
		var container varTestStruct
		if err := PerformSSZCheck(objBytes, &container, yamlPath, valid); err != nil {
			t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
		}
		if err := PerformRootCheck(container, yamlPath, valid); err != nil {
			t.Fatalf("Could not perform root check for case %s: %v", testName, err)
		}
	}
}

func runUintSSZTestCase(t *testing.T, objBytes []byte, yamlPath string, testName string, valid bool) error {
	switch testName {
	case "uint_8":
		var result uint8
		if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
			t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
		}
		if err := PerformRootCheck(result, yamlPath, valid); err != nil {
			t.Fatalf("Could not perform root check for case %s: %v", testName, err)
		}
	case "uint_16":
		var result uint16
		if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
			t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
		}
		if err := PerformRootCheck(result, yamlPath, valid); err != nil {
			t.Fatalf("Could not perform root check for case %s: %v", testName, err)
		}
	case "uint_32":
		var result uint32
		if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
			t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
		}
		if err := PerformRootCheck(result, yamlPath, valid); err != nil {
			t.Fatalf("Could not perform root check for case %s: %v", testName, err)
		}
	case "uint_64":
		var result uint64
		if err := PerformSSZCheck(objBytes, &result, yamlPath, valid); err != nil {
			t.Fatalf("could not perform ssz check for case %s: %v", testName, err)
		}
		if err := PerformRootCheck(result, yamlPath, valid); err != nil {
			t.Fatalf("Could not perform root check for case %s: %v", testName, err)
		}
	case "uint_128":
		var resultBytes []byte
		var result = new(big.Int)
		err := ssz.Unmarshal(objBytes, &resultBytes)
		if valid {
			if err != nil {
				t.Fatal(err)
			}
			result.SetBytes(resultBytes)
			encoded, err := ssz.Marshal(result.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(objBytes, encoded) {
				t.Fatalf("expected %v, received %v", objBytes, encoded)
			}
		}

		// TODO add conditions from PerformSSZCheck here
	case "uint_256":
		var resultBytes []byte
		var result = new(big.Int)
		err := ssz.Unmarshal(objBytes, &resultBytes)
		if valid {
			if err != nil {
				t.Fatal(err)
			}
			result.SetBytes(resultBytes)
			encoded, err := ssz.Marshal(result.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(objBytes, encoded) {
				t.Fatalf("expected %v, received %v", objBytes, encoded)
			}
		}

		// TODO add conditions from PerformSSZCheck here
	}
	return nil
}

func PerformSSZCheck(objBytes []byte, val interface{}, yamlPath string, valid bool) error {
	err := ssz.Unmarshal(objBytes, val)
	if valid {
		if err != nil {
			return fmt.Errorf("test case shouldn't have failed: %v", err)
		}

		encoded, err := ssz.Marshal(val)
		if err != nil {
			return fmt.Errorf("failed to marshal: %v", err)
		}
		if !bytes.Equal(objBytes, encoded) {
			return fmt.Errorf("incorrect encoding/decoding, expected %#x, received %#x", objBytes, encoded)
		}
	} else if err == nil {
		return fmt.Errorf("expected error, received nil: %v", val)
	}
	return nil
}

func PerformRootCheck(val interface{}, yamlPath string, valid bool) error {
	if !valid {
		return nil
	}

	yamlBytes, err := FileBytesOrDie(yamlPath)
	if err != nil {
		return err
	}

	var roots sszRoots
	if err := yaml.Unmarshal(yamlBytes, &roots); err != nil {
		return fmt.Errorf("failed to unmarshal: %v", err)
	}

	correctRoot, err := hex.DecodeString(roots.Root[2:])
	if err != nil {
		return fmt.Errorf("failed to decode: %v", err)
	}

	root, err := ssz.HashTreeRoot(val)
	if err != nil {
		return fmt.Errorf("failed to get hashtreeroot: %v", err)
	}

	if !bytes.Equal(correctRoot, root[:]) {
		return fmt.Errorf("failed hash tree root check, expected %#x, received %#x", correctRoot, root)
	}

	if roots.SigningRoot != "" {
		correctRoot, err = hex.DecodeString(roots.SigningRoot[2:])
		if err != nil {
			return fmt.Errorf("failed to decode: %v", err)
		}

		root, err = ssz.SigningRoot(val)
		if err != nil {
			return fmt.Errorf("failed to get hashtreeroot: %v", err)
		}

		if !bytes.Equal(correctRoot, root[:]) {
			return fmt.Errorf("failed signing root check, expected %#x, received %#x", correctRoot, root)
		}
	}
	return nil
}

func PerformBitlistRootCheck(bitlist bitfield.Bitlist, size uint64, yamlPath string, valid bool) error {
	yamlBytes, err := FileBytesOrDie(yamlPath)
	if err != nil {
		return err
	}

	var roots sszRoots
	if err := yaml.Unmarshal(yamlBytes, &roots); err != nil {
		return fmt.Errorf("failed to unmarshal: %v", err)
	}

	correctRoot, err := hex.DecodeString(roots.Root[2:])
	if err != nil {
		return fmt.Errorf("failed to decode: %v", err)
	}

	root, err := ssz.HashTreeRootBitlist(bitlist, size)
	if err != nil {
		return fmt.Errorf("failed to get hashtreeroot: %v", err)
	}

	if !bytes.Equal(correctRoot, root[:]) {
		return fmt.Errorf("failed hash tree root check, expected %#x, received %#x", correctRoot, root)
	}
	return nil
}

func PerformListRootCheck(val interface{}, size uint64, yamlPath string, valid bool) error {
	if !valid {
		return nil
	}

	yamlBytes, err := FileBytesOrDie(yamlPath)
	if err != nil {
		return err
	}

	var roots sszRoots
	if err := yaml.Unmarshal(yamlBytes, &roots); err != nil {
		return fmt.Errorf("failed to unmarshal: %v", err)
	}

	correctRoot, err := hex.DecodeString(roots.Root[2:])
	if err != nil {
		return fmt.Errorf("failed to decode: %v", err)
	}

	root, err := ssz.HashTreeRootWithCapacity(val, size)
	if err != nil {
		return fmt.Errorf("failed to get hashtreeroot: %v", err)
	}

	if !bytes.Equal(correctRoot, root[:]) {
		return fmt.Errorf("failed hash tree root check, expected %#x, received %#x", correctRoot, root)
	}

	if roots.SigningRoot != "" {
		correctRoot, err = hex.DecodeString(roots.SigningRoot[2:])
		if err != nil {
			return fmt.Errorf("failed to decode: %v", err)
		}

		root, err = ssz.SigningRoot(val)
		if err != nil {
			return fmt.Errorf("failed to get hashtreeroot: %v", err)
		}

		if !bytes.Equal(correctRoot, root[:]) {
			return fmt.Errorf("failed signing root check, expected %#x, received %#x", correctRoot, root)
		}
	}
	return nil
}

func FileBytesOrDie(path string) ([]byte, error) {
	file, err := bazel.Runfile(path)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to marshal: %v", err)
	}
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to marshal: %v", err)
	}
	return bytes, nil
}
