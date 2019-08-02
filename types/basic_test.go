package types

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestBasicSSZ_Root(t *testing.T) {
	type args struct {
		v interface{}
	}
	items := [120000]uint64{}
	for i := 0; i < 120000; i++ {
		items[i] = uint64(i)
	}
	val := reflect.ValueOf(items)
	typ := reflect.TypeOf(items)
	factory, _ := SSZFactory(val, typ)

	start := time.Now()
	got, err := factory.Root(val, typ)
	if err != nil {
		t.Fatal(err)
	}
	end := time.Now()
	fmt.Printf("%#x and time %v\n", got, end.Sub(start))
	start = time.Now()
	got, err = factory.Root(val, typ)
	if err != nil {
		t.Fatal(err)
	}
	end = time.Now()
	fmt.Printf("%#x and time %v\n", got, end.Sub(start))

	items[200] = uint64(302)
	val = reflect.ValueOf(items)
	typ = reflect.TypeOf(items)
	start = time.Now()
	got, err = factory.Root(val, typ)
	if err != nil {
		t.Fatal(err)
	}
	end = time.Now()
	fmt.Printf("%#x and time %v and added one more item\n", got, end.Sub(start))
}
