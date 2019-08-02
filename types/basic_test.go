package types

import (
	"reflect"
	"testing"
	"time"
)

func TestBasicSSZ_Root(t *testing.T) {
	factory := NewBasicSSZ()
	type args struct {
		v interface{}
	}
	items := []uint64{}
	for i := 0; i < 120000; i++ {
		items = append(items, uint64(i))
	}
	val := reflect.ValueOf(items)
	typ := reflect.TypeOf(items)

	start := time.Now()
	got, err := factory.Root(val, typ)
	if err != nil {
		t.Fatal(err)
	}
	end := time.Now()
	t.Logf("%#x and time %v", got, end.Sub(start))
	start = time.Now()
	got, err = factory.Root(val, typ)
	if err != nil {
		t.Fatal(err)
	}
	end = time.Now()
	t.Logf("%#x and time %v", got, end.Sub(start))

	items = append(items, uint64(320323))
	val = reflect.ValueOf(items)
	typ = reflect.TypeOf(items)
	start = time.Now()
	got, err = factory.Root(val, typ)
	if err != nil {
		t.Fatal(err)
	}
	end = time.Now()
	t.Logf("%#x and time %v and added one more item", got, end.Sub(start))
}
