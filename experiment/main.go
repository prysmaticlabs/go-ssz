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

const BytesPerChunk = 32

type marshaler func(val reflect.Value, buf []byte, startOffset uint64) (uint64, error)

type BeaconState struct {
	BlockRoots [65536][32]byte
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
		BlockRoots: [65536][32]byte{},
	}
	for i := 0; i < len(bs.BlockRoots); i++ {
		bs.BlockRoots[i] = [32]byte{1, 2, 3}
	}
	start := time.Now()
	enc, err := ssz.HashTreeRoot(bs)
	if err != nil {
		panic(err)
	}
	end := time.Now()
	log.Printf("%#x in %v", enc, end.Sub(start))
	// Change 2 elements.
	bs.BlockRoots[len(bs.BlockRoots)-1] = [32]byte{4, 5, 6}
	bs.BlockRoots[1] = [32]byte{7, 8, 9}
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
