package types

import (
	"reflect"
	"strconv"
	"testing"
)

func BenchmarkBasicArrayRoot_WithCache(b *testing.B) {
	b.StopTimer()
	items := [1000][]byte{}
	for i := 0; i < len(items); i++ {
		items[i] = make([]byte, 32)
		copy(items[i], []byte(strconv.Itoa(i)))
	}
	ss := newBasicArraySSZ()
	v := reflect.ValueOf(items)
	typ := reflect.TypeOf([1000][32]byte{})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ss.Root(v, typ, 0); err != nil {
			b.Fatal(err)
		}
	}
}
