package main

import (
	"encoding/binary"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/prysmaticlabs/go-ssz"
	pb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
)

type marshaler func(val reflect.Value, buf []byte, startOffset uint64) (uint64, error)

type BeaconState struct {
	BlockRoots       [][]byte `ssz-size:"65536,32"`
	ActiveIndexRoots [][]byte `ssz-size:"2048,32"`
}

var (
	typesMap = make(map[reflect.Type]marshaler)
)

func init() {
	check := pb.Checkpoint{
		Epoch: 1,
		Root:  make([]byte, 32),
	}
	copy(check.Root, "12345678910")
	val := reflect.ValueOf(check)
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		if strings.Contains(typ.Field(i).Name, "XXX_") {
			continue
		}
		if val.Field(i).Type().Kind() == reflect.Uint64 {
			typesMap[val.Field(i).Type()] = marshalUint64
		} else if val.Field(i).Type().Kind() == reflect.Slice {
			typesMap[val.Field(i).Type()] = marshalByteArray
		}
	}
}

func main() {
	bs := BeaconState{
		BlockRoots:       make([][]byte, 65536),
		ActiveIndexRoots: make([][]byte, 2048),
	}
	for i := 0; i < len(bs.BlockRoots); i++ {
		item := [32]byte{1, 2, 3}
		bs.BlockRoots[i] = item[:]
	}
	for i := 0; i < len(bs.ActiveIndexRoots); i++ {
		item := [32]byte{4, 5, 6}
		bs.ActiveIndexRoots[i] = item[:]
	}
	start := time.Now()
	enc, err := ssz.HashTreeRoot(bs)
	if err != nil {
		panic(err)
	}
	end := time.Now()
	log.Printf("%#x in %v", enc, end.Sub(start))
	start = time.Now()
	enc, err = ssz.HashTreeRoot(bs)
	if err != nil {
		panic(err)
	}
	end = time.Now()
	log.Printf("%#x in %v", enc, end.Sub(start))
	// Change 2 elements.
	newItem1 := [32]byte{4, 5, 6}
	newItem2 := [32]byte{7, 8, 9}
	bs.BlockRoots[len(bs.BlockRoots)-1] = newItem1[:]
	bs.BlockRoots[1] = newItem2[:]
	// Change 1 element.
	newItem := [32]byte{1, 1, 1}
	bs.ActiveIndexRoots[1] = newItem[:]
	start = time.Now()
	enc, err = ssz.HashTreeRoot(bs)
	if err != nil {
		panic(err)
	}
	end = time.Now()
	log.Printf("%#x in %v", enc, end.Sub(start))
}

func fastMarshal(check pb.Checkpoint) ([]byte, error) {
	val := reflect.ValueOf(check)
	typ := val.Type()
	buf := make([]byte, 40)
	position := uint64(0)
	var err error
	for i := 0; i < typ.NumField(); i++ {
		if strings.Contains(typ.Field(i).Name, "XXX_") {
			continue
		}
		ff := typesMap[val.Field(i).Type()]
		position, err = ff(val.Field(i), buf, position)
		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func marshalUint64(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	binary.LittleEndian.PutUint64(buf[startOffset:], val.Interface().(uint64))
	return startOffset + 8, nil
}

func marshalByteArray(val reflect.Value, buf []byte, startOffset uint64) (uint64, error) {
	copy(buf[startOffset:], val.Bytes())
	return startOffset + uint64(val.Len()), nil
}
