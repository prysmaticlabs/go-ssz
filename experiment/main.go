package main

import (
	"encoding/binary"
	"errors"

	"github.com/minio/sha256-simd"
	"github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/merkle"
	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
)

func main() {

}

func stateRoot(state *pb.BeaconState) {
	// There are 20 fields in the beacon state.
	fieldRoots := [20][32]byte{}

	// Do the genesis time:
	genesisBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(genesisBuf, state.GenesisTime)
	fieldRoots[0] = bytesutil.ToBytes32(genesisBuf)
	// Do the slot:
	slotBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(slotBuf, state.Slot)
	fieldRoots[1] = bytesutil.ToBytes32(slotBuf)

	// Handle the fork data:
	forkRoots := make([][]byte, 3)
	inter := bytesutil.ToBytes32(state.Fork.PreviousVersion)
	forkRoots[0] = inter[:]
	inter = bytesutil.ToBytes32(state.Fork.CurrentVersion)
	forkRoots[1] = inter[:]
	forkEpochBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(forkEpochBuf, state.Fork.Epoch)
	inter = bytesutil.ToBytes32(forkEpochBuf)
	forkRoots[2] = inter[:]
	forkRoot, err := bitwiseMerkleize(forkRoots, 3, 3)
	if err != nil {
		panic(err)
	}
	fieldRoots[2] = forkRoot

	// Handle the beacon block header:
}

// Given ordered BYTES_PER_CHUNK-byte chunks, if necessary utilize zero chunks so that the
// number of chunks is a power of two, Merkleize the chunks, and return the root.
// Note that merkleize on a single chunk is simply that chunk, i.e. the identity
// when the number of chunks is one.
func bitwiseMerkleize(chunks [][]byte, count uint64, limit uint64) ([32]byte, error) {
	if count > limit {
		return [32]byte{}, errors.New("merkleizing list that is too large, over limit")
	}
	hasher := htr.HashFn(hash)
	leafIndexer := func(i uint64) []byte {
		return chunks[i]
	}
	return merkle.Merkleize(hasher, count, limit, leafIndexer), nil
}

// hash defines a function that returns the sha256 hash of the data passed in.
func hash(data []byte) [32]byte {
	return sha256.Sum256(data)
}
