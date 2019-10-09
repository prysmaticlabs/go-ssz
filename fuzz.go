// +build gofuzz

package ssz

func Fuzz(data []byte) int {
	type Base struct {
		F1 bool
		F2 uint8
		F3 uint16
		F4 uint32
		F5 uint64
		F6 []byte
		F7 []Base
		F8 *Base
	}

	type T struct {
		F9 Base
	}

	var v T
	if err := Unmarshal(data, &v); err != nil {
		return 0
	}
	return 1
}
