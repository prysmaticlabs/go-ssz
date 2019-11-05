package spectests

import (
	"io/ioutil"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/prysmaticlabs/go-ssz"
)

type SszBenchmarkState struct {
	Value      minimalBeaconState `json:"value"`
	Serialized []byte             `json:"serialized"`
	Root       []byte             `json:"root" ssz:"size=32"`
}

type SszBenchmarkBlock struct {
	Value       minimalBlock `json:"value"`
	Serialized  []byte       `json:"serialized"`
	Root        []byte       `json:"root" ssz:"size=32"`
	SigningRoot []byte       `json:"signing_root" ssz:"size=96"`
}

func BenchmarkBeaconBlock_Marshal(b *testing.B) {
	b.StopTimer()
	s := &SszBenchmarkBlock{}
	populateStructFromYaml(b, "./yaml/ssz_single_block.yaml", s)
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		if _, err := ssz.Marshal(s.Value); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBeaconBlock_Unmarshal(b *testing.B) {
	b.StopTimer()
	s := &SszBenchmarkBlock{}
	populateStructFromYaml(b, "./yaml/ssz_single_block.yaml", s)
	encoded, err := ssz.Marshal(s.Value)
	if err != nil {
		b.Fatal(err)
	}
	var target minimalBlock
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		if err := ssz.Unmarshal(encoded, &target); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBeaconBlock_HashTreeRoot(b *testing.B) {
	b.StopTimer()
	s := &SszBenchmarkBlock{}
	populateStructFromYaml(b, "./yaml/ssz_single_block.yaml", s)
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		if _, err := ssz.HashTreeRoot(s.Value); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBeaconState_Marshal(b *testing.B) {
	b.StopTimer()
	s := &SszBenchmarkState{}
	populateStructFromYaml(b, "./yaml/ssz_single_state.yaml", s)
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		if _, err := ssz.Marshal(s.Value); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBeaconState_Unmarshal(b *testing.B) {
	b.StopTimer()
	s := &SszBenchmarkState{}
	populateStructFromYaml(b, "./yaml/ssz_single_state.yaml", s)
	encoded, err := ssz.Marshal(s.Value)
	if err != nil {
		b.Fatal(err)
	}
	var target minimalBeaconState
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		if err := ssz.Unmarshal(encoded, &target); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBeaconState_HashTreeRoot(b *testing.B) {
	b.StopTimer()
	s := &SszBenchmarkState{}
	populateStructFromYaml(b, "./yaml/ssz_single_state.yaml", s)
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		if _, err := ssz.HashTreeRoot(s.Value); err != nil {
			b.Fatal(err)
		}
	}
}

func populateStructFromYaml(t testing.TB, fPath string, val interface{}) {
	yamlFile, err := ioutil.ReadFile(fPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(yamlFile, val); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
}
