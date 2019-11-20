package main

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/minio/sha256-simd"
	"github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/merkle"
	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
)

const BytesPerChunk = 32

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
	blockHeaderRoots := make([][]byte, 5)
	headerSlotBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(headerSlotBuf, state.LatestBlockHeader.Slot)
	inter = bytesutil.ToBytes32(headerSlotBuf)
	blockHeaderRoots[0] = inter[:]
	blockHeaderRoots[1] = state.LatestBlockHeader.ParentRoot
	blockHeaderRoots[2] = state.LatestBlockHeader.StateRoot
	blockHeaderRoots[4] = state.LatestBlockHeader.BodyRoot
	signatureChunks, err := pack([][]byte{state.LatestBlockHeader.Signature})
	if err != nil {
		panic(err)
	}
	sigRoot, err := bitwiseMerkleize(signatureChunks, uint64(len(signatureChunks)), uint64(len(signatureChunks)))
	if err != nil {
		panic(err)
	}
	blockHeaderRoots[5] = sigRoot[:]
	headerRoot, err := bitwiseMerkleize(blockHeaderRoots, 5, 5)
	if err != nil {
		panic(err)
	}
	fieldRoots[3] = headerRoot

	// Handle the block roots:
	fieldRoots[4] = merkleize(state.BlockRoots)
	// Handle the state roots:
	fieldRoots[5] = merkleize(state.StateRoots)

	// Handle the historical roots:
	historicalRootsBuf := new(bytes.Buffer)
	if err := binary.Write(historicalRootsBuf, binary.LittleEndian, uint64(len(state.HistoricalRoots))); err != nil {
		panic(err)
	}
	historicalRootsOutput := make([]byte, 32)
	copy(historicalRootsOutput, historicalRootsBuf.Bytes())
	merkleRoot, err := bitwiseMerkleize(state.HistoricalRoots, uint64(len(state.HistoricalRoots)), 16777216)
	if err != nil {
		panic(err)
	}
	fieldRoots[6] = mixInLength(merkleRoot, historicalRootsOutput)

	// Handle the eth1 data:
	fieldRoots[7] = eth1Root(state.Eth1Data)

	// Handle eth1 data votes:
	eth1VotesRoots := make([][]byte, 1024)
	for i := 0; i < len(eth1VotesRoots); i++ {
		inter = eth1Root(state.Eth1DataVotes[i])
		eth1VotesRoots[i] = inter[:]
	}
	eth1VotesRootsRoot, err := bitwiseMerkleize(eth1VotesRoots, uint64(len(eth1VotesRoots)), uint64(len(eth1VotesRoots)))
	if err != nil {
		panic(err)
	}
	fieldRoots[8] = eth1VotesRootsRoot

	// Handle eth1 deposit index:
	eth1DepositIndexBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(eth1DepositIndexBuf, state.Eth1DepositIndex)
	inter = bytesutil.ToBytes32(eth1DepositIndexBuf)
	fieldRoots[9] = inter

	// Handle the validator registry:
	validatorsRoots := make([][]byte, 0)
	for i := 0; i < len(state.Validators); i++ {
		inter = validatorRoot(state.Validators[i])
		validatorsRoots = append(validatorsRoots, inter[:])
	}
	validatorsRootsRoot, err := bitwiseMerkleize(validatorsRoots, uint64(len(validatorsRoots)), uint64(1099511627776))
	if err != nil {
		panic(err)
	}
	fieldRoots[10] = validatorsRootsRoot

	// Handle the validator balances:
	balancesRoots := make([][]byte, 0)
	for i := 0; i < len(state.Balances); i++ {
		balanceBuf := make([]byte, 8)
		binary.LittleEndian.PutUint64(balanceBuf, state.Balances[i])
		inter = bytesutil.ToBytes32(balanceBuf)
		balancesRoots = append(balancesRoots, inter[:])
	}
	balancesRootsRoot, err := bitwiseMerkleize(balancesRoots, uint64(len(balancesRoots)), uint64(1099511627776))
	if err != nil {
		panic(err)
	}
	fieldRoots[11] = balancesRootsRoot

	// Handle the randao mixes:
	fieldRoots[12] = merkleize(state.RandaoMixes)

	// Handle the slashings:
	slashingRoots := make([][]byte, 8192)
	for i := 0; i < len(slashingRoots); i++ {
		slashingRoot := make([]byte, 8)
		binary.LittleEndian.PutUint64(slashingRoot, state.Slashings[i])
		inter = bytesutil.ToBytes32(slashingRoot)
		slashingRoots[i] = inter[:]
	}
	slashingRootsRoot, err := bitwiseMerkleize(slashingRoots, uint64(len(slashingRoots)), uint64(len(slashingRoots)))
	if err != nil {
		panic(err)
	}
	fieldRoots[13] = slashingRootsRoot

	// Handle the previous epoch attestations 14:
	prevAttsLenBuf := new(bytes.Buffer)
	if err := binary.Write(prevAttsLenBuf, binary.LittleEndian, uint64(4096)); err != nil {
		panic(err)
	}
	prevAttsLenRoot := make([]byte, 32)
	copy(prevAttsLenRoot, prevAttsLenBuf.Bytes())
	prevAttsRoots := make([][]byte, 0)
	for i := 0; i < len(state.PreviousEpochAttestations); i++ {
		inter = pendingAttestationRoot(state.PreviousEpochAttestations[i])
		prevAttsRoots = append(prevAttsRoots, inter[:])
	}
	prevAttsRootsRoot, err := bitwiseMerkleize(prevAttsRoots, uint64(len(prevAttsRoots)), 4096)
	if err != nil {
		panic(err)
	}
	fieldRoots[14] = mixInLength(prevAttsRootsRoot, prevAttsLenRoot)

	// Handle the current epoch attestations 15:
	currAttsLenBuf := new(bytes.Buffer)
	if err := binary.Write(currAttsLenBuf, binary.LittleEndian, uint64(4096)); err != nil {
		panic(err)
	}
	currAttsLenRoot := make([]byte, 32)
	copy(currAttsLenRoot, currAttsLenBuf.Bytes())
	currAttsRoots := make([][]byte, 0)
	for i := 0; i < len(state.CurrentEpochAttestations); i++ {
		inter = pendingAttestationRoot(state.CurrentEpochAttestations[i])
		currAttsRoots = append(currAttsRoots, inter[:])
	}
	currAttsRootsRoot, err := bitwiseMerkleize(currAttsRoots, uint64(len(currAttsRoots)), 4096)
	if err != nil {
		panic(err)
	}
	fieldRoots[15] = mixInLength(currAttsRootsRoot, currAttsLenRoot)

	// Handle the justification bits 16:

	// Handle the previous justified checkpoint 17:
	fieldRoots[17] = checkpointRoot(state.PreviousJustifiedCheckpoint)
	// Handle the current justified checkpoint 18:
	fieldRoots[18] = checkpointRoot(state.CurrentJustifiedCheckpoint)
	// Handle the finalized checkpoint 19:
	fieldRoots[19] = checkpointRoot(state.FinalizedCheckpoint)
}

func pendingAttestationRoot(att *pb.PendingAttestation) [32]byte {
	fieldRoots := make([][]byte, 4)

	// Bitfield.

	// Attestation data.

	// Inclusion delay.
	inclusionBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(inclusionBuf, att.InclusionDelay)
	inter := bytesutil.ToBytes32(inclusionBuf)
	fieldRoots[2] = inter[:]

	// Proposer index.
	proposerBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(proposerBuf, att.ProposerIndex)
	inter = bytesutil.ToBytes32(proposerBuf)
	fieldRoots[3] = inter[:]

	root, err := bitwiseMerkleize(fieldRoots, 4, 4)
	if err != nil {
		panic(err)
	}
	return root
}

func validatorRoot(validator *ethpb.Validator) [32]byte {
	fieldRoots := make([][]byte, 8)

	// Public key.
	pubKeyChunks, err := pack([][]byte{validator.PublicKey})
	if err != nil {
		panic(err)
	}
	pubKeyRoot, err := bitwiseMerkleize(pubKeyChunks, uint64(len(pubKeyChunks)), uint64(len(pubKeyChunks)))
	if err != nil {
		panic(err)
	}
	fieldRoots[0] = pubKeyRoot[:]

	// Withdrawal credentials.
	fieldRoots[1] = validator.WithdrawalCredentials

	// Effective balance.
	effectiveBalanceBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(effectiveBalanceBuf, validator.EffectiveBalance)
	inter := bytesutil.ToBytes32(effectiveBalanceBuf)
	fieldRoots[2] = inter[:]

	// Slashed.
	slashBuf := make([]byte, 1)
	if validator.Slashed {
		slashBuf[0] = uint8(1)
	} else {
		slashBuf[0] = uint8(0)
	}
	inter = bytesutil.ToBytes32(slashBuf)
	fieldRoots[3] = inter[:]

	// Activation eligibility epoch.
	activationEligibilityBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(activationEligibilityBuf, validator.ActivationEligibilityEpoch)
	inter = bytesutil.ToBytes32(activationEligibilityBuf)
	fieldRoots[4] = inter[:]

	// Activation epoch.
	activationBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(activationBuf, validator.ActivationEpoch)
	inter = bytesutil.ToBytes32(activationBuf)
	fieldRoots[5] = inter[:]

	// Exit epoch.
	exitBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(exitBuf, validator.ExitEpoch)
	inter = bytesutil.ToBytes32(exitBuf)
	fieldRoots[6] = inter[:]

	// Withdrawable epoch.
	withdrawalBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(withdrawalBuf, validator.WithdrawableEpoch)
	inter = bytesutil.ToBytes32(withdrawalBuf)
	fieldRoots[7] = inter[:]

	root, err := bitwiseMerkleize(fieldRoots, 3, 3)
	if err != nil {
		panic(err)
	}
	return root
}

func eth1Root(eth1Data *ethpb.Eth1Data) [32]byte {
	fieldRoots := make([][]byte, 3)
	fieldRoots[0] = eth1Data.DepositRoot
	eth1DataCountBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(eth1DataCountBuf, eth1Data.DepositCount)
	inter := bytesutil.ToBytes32(eth1DataCountBuf)
	fieldRoots[1] = inter[:]
	fieldRoots[2] = eth1Data.BlockHash
	eth1DataRoot, err := bitwiseMerkleize(fieldRoots, 3, 3)
	if err != nil {
		panic(err)
	}
	return eth1DataRoot
}

func checkpointRoot(checkpoint *ethpb.Checkpoint) [32]byte {
	fieldRoots := make([][]byte, 2)
	epochBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(epochBuf, checkpoint.Epoch)
	inter := bytesutil.ToBytes32(epochBuf)
	fieldRoots[0] = inter[:]
	fieldRoots[1] = checkpoint.Root
	checkpointRoot, err := bitwiseMerkleize(fieldRoots, 2, 2)
	if err != nil {
		panic(err)
	}
	return checkpointRoot
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

func pack(serializedItems [][]byte) ([][]byte, error) {
	areAllEmpty := true
	for _, item := range serializedItems {
		if !bytes.Equal(item, []byte{}) {
			areAllEmpty = false
			break
		}
	}
	// If there are no items, we return an empty chunk.
	if len(serializedItems) == 0 || areAllEmpty {
		emptyChunk := make([]byte, BytesPerChunk)
		return [][]byte{emptyChunk}, nil
	} else if len(serializedItems[0]) == BytesPerChunk {
		// If each item has exactly BYTES_PER_CHUNK length, we return the list of serialized items.
		return serializedItems, nil
	}
	// We flatten the list in order to pack its items into byte chunks correctly.
	orderedItems := []byte{}
	for _, item := range serializedItems {
		orderedItems = append(orderedItems, item...)
	}
	numItems := len(orderedItems)
	chunks := [][]byte{}
	for i := 0; i < numItems; i += BytesPerChunk {
		j := i + BytesPerChunk
		// We create our upper bound index of the chunk, if it is greater than numItems,
		// we set it as numItems itself.
		if j > numItems {
			j = numItems
		}
		// We create chunks from the list of items based on the
		// indices determined above.
		chunks = append(chunks, orderedItems[i:j])
	}
	// Right-pad the last chunk with zero bytes if it does not
	// have length BytesPerChunk.
	lastChunk := chunks[len(chunks)-1]
	for len(lastChunk) < BytesPerChunk {
		lastChunk = append(lastChunk, 0)
	}
	chunks[len(chunks)-1] = lastChunk
	return chunks, nil
}

func merkleize(chunks [][]byte) [32]byte {
	if len(chunks) == 1 {
		var root [32]byte
		copy(root[:], chunks[0])
		return root
	}
	for !isPowerOf2(len(chunks)) {
		chunks = append(chunks, make([]byte, BytesPerChunk))
	}
	hashLayer := chunks
	// We keep track of the hash layers of a Merkle trie until we reach
	// the top layer of length 1, which contains the single root element.
	//        [Root]      -> Top layer has length 1.
	//    [E]       [F]   -> This layer has length 2.
	// [A]  [B]  [C]  [D] -> The bottom layer has length 4 (needs to be a power of two).
	i := 1
	for len(hashLayer) > 1 {
		layer := [][]byte{}
		for i := 0; i < len(hashLayer); i += 2 {
			hashedChunk := hash(append(hashLayer[i], hashLayer[i+1]...))
			layer = append(layer, hashedChunk[:])
		}
		hashLayer = layer
		i++
	}
	var root [32]byte
	copy(root[:], hashLayer[0])
	return root
}

func isPowerOf2(n int) bool {
	return n != 0 && (n&(n-1)) == 0
}

func mixInLength(root [32]byte, length []byte) [32]byte {
	var hash [32]byte
	h := sha256.New()
	h.Write(root[:])
	h.Write(length)
	// The hash interface never returns an error, for that reason
	// we are not handling the error below. For reference, it is
	// stated here https://golang.org/pkg/hash/#Hash
	// #nosec G104
	h.Sum(hash[:0])
	return hash
}
