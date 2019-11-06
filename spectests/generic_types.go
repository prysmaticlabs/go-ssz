package spectests

type sszRoots struct {
	Root        string `yaml:"root"`
	SigningRoot string `yaml:"signing_root"`
}

type singleFieldStruct struct {
	A byte
}

type smallTestStruct struct {
	A uint16
	B uint16
}

type fixedTestStruct struct {
	A uint8
	B uint64
	C uint32
}

type varTestStruct struct {
	A uint16
	B []uint16 `ssz-max:"1024"`
	C uint8
}

type complexTestStruct struct {
	A uint16
	B []uint16 `ssz-max:"128"`
	C uint8
	D []byte `ssz-max:"256"`
	E varTestStruct
	F []fixedTestStruct `ssz-size:"4"`
	G []varTestStruct   `ssz-size:"2"`
}
