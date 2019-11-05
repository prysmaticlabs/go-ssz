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
	A uint16            // size = 2
	B []uint16          `ssz-max:"128"` // size = 2 + 4 = 6
	C uint8             // size = 1
	D []byte            `ssz-size:"256"` // size = 256
	E varTestStruct     // size 2 + 6 + 1 = 9 + 4 = 13
	F []fixedTestStruct `ssz-size:"4"` // size 13 * 4 = 52
	G []varTestStruct   `ssz-size:"2"` // size 14 + 4 =  18
}
