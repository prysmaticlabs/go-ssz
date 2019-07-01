package ssz

func DeepEqual(val1 interface{}, val2 interface{}) bool {
	return val1.(uint64) == val2.(uint64)
}
