package spectests

import (
	"github.com/prysmaticlabs/go-bitfield"
)

type minimalFork struct {
	PreviousVersion []byte `json:"previous_version" ssz-size:"4"`
	CurrentVersion  []byte `json:"current_version" ssz-size:"4"`
	Epoch           uint64 `json:"epoch"`
}

type minimalCheckpoint struct {
	Epoch uint64 `json:"epoch"`
	Root  []byte `json:"root" ssz-size:"32"`
}

type minimalValidator struct {
	Pubkey                     []byte `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials      []byte `json:"withdrawal_credentials" ssz-size:"32"`
	EffectiveBalance           uint64 `json:"effective_balance"`
	Slashed                    bool   `json:"slashed"`
	ActivationEligibilityEpoch uint64 `json:"activation_eligibility_epoch"`
	ActivationEpoch            uint64 `json:"activation_epoch"`
	ExitEpoch                  uint64 `json:"exit_epoch"`
	WithdrawableEpoch          uint64 `json:"withdrawable_epoch"`
}

type minimalAttestationData struct {
	Slot            uint64
	Index           uint64
	BeaconBlockRoot []byte            `json:"beacon_block_root" ssz-size:"32"`
	Source          minimalCheckpoint `json:"source"`
	Target          minimalCheckpoint `json:"target"`
}

type minimalAttestationAndCustodyBit struct {
	Data       minimalAttestationData `json:"data"`
	CustodyBit bool                   `json:"custody_bit"`
}

type minimalIndexedAttestation struct {
	CustodyBit0Indices []uint64               `json:"custody_bit_0_indices" ssz-max:"2048"`
	CustodyBit1Indices []uint64               `json:"custody_bit_1_indices" ssz-max:"2048"`
	Data               minimalAttestationData `json:"data"`
	Signature          []byte                 `json:"signature" ssz-size:"96"`
}

type minimalPendingAttestation struct {
	AggregationBits bitfield.Bitlist       `json:"aggregation_bits" ssz-max:"2048"`
	Data            minimalAttestationData `json:"data"`
	InclusionDelay  uint64                 `json:"inclusion_delay"`
	ProposerIndex   uint64                 `json:"proposer_index"`
}

type minimalEth1Data struct {
	DepositRoot  []byte `json:"deposit_root" ssz-size:"32"`
	DepositCount uint64 `json:"deposit_count"`
	BlockHash    []byte `json:"block_hash" ssz-size:"32"`
}

type minimalHistoricalBatch struct {
	BlockRoots [][]byte `json:"block_roots" ssz-size:"64,32"`
	StateRoots [][]byte `json:"state_roots" ssz-size:"64,32"`
}

type minimalDepositData struct {
	Pubkey                []byte `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials []byte `json:"withdrawal_credentials" ssz-size:"32"`
	Amount                uint64 `json:"amount"`
	Signature             []byte `json:"signature" ssz-size:"96"`
}

type minimalBlockHeader struct {
	Slot       uint64 `json:"slot"`
	ParentRoot []byte `json:"parent_root" ssz-size:"32"`
	StateRoot  []byte `json:"state_root" ssz-size:"32"`
	BodyRoot   []byte `json:"body_root" ssz-size:"32"`
	Signature  []byte `json:"signature" ssz-size:"96"`
}

type minimalProposerSlashing struct {
	ProposerIndex uint64             `json:"proposer_index"`
	Header1       minimalBlockHeader `json:"header_1"`
	Header2       minimalBlockHeader `json:"header_2"`
}

type minimalAttesterSlashing struct {
	Attestation1 minimalIndexedAttestation `json:"attestation_1"`
	Attestation2 minimalIndexedAttestation `json:"attestation_2"`
}

type minimalAttestation struct {
	AggregationBits bitfield.Bitlist       `json:"aggregation_bits" ssz-max:"2048"`
	Data            minimalAttestationData `json:"data"`
	CustodyBits     bitfield.Bitlist       `json:"custody_bits" ssz-max:"2048"`
	Signature       []byte                 `json:"signature" ssz-size:"96"`
}

type minimalDeposit struct {
	Proof [][]byte           `json:"proof" ssz-size:"33,32"`
	Data  minimalDepositData `json:"data"`
}

type minimalVoluntaryExit struct {
	Epoch          uint64 `json:"epoch"`
	ValidatorIndex uint64 `json:"validator_index"`
	Signature      []byte `json:"signature" ssz-size:"96"`
}

type minimalBlockBody struct {
	RandaoReveal      []byte                    `json:"randao_reveal" ssz-size:"96"`
	Eth1Data          minimalEth1Data           `json:"eth1_data"`
	Graffiti          []byte                    `json:"graffiti" ssz-size:"32"`
	ProposerSlashings []minimalProposerSlashing `json:"proposer_slashings" ssz-max:"16"`
	AttesterSlashings []minimalAttesterSlashing `json:"attester_slashings" ssz-max:"1"`
	Attestations      []minimalAttestation      `json:"attestations" ssz-max:"128"`
	Deposits          []minimalDeposit          `json:"deposits" ssz-max:"16"`
	VoluntaryExits    []minimalVoluntaryExit    `json:"voluntary_exits" ssz-max:"16"`
}

type minimalBlock struct {
	Slot       uint64           `json:"slot"`
	ParentRoot []byte           `json:"parent_root" ssz-size:"32"`
	StateRoot  []byte           `json:"state_root" ssz-size:"32"`
	Body       minimalBlockBody `json:"body"`
	Signature  []byte           `json:"signature" ssz-size:"96"`
}

type minimalAggregateAndProof struct {
	Index          uint64
	SelectionProof []byte `ssz-size:"96"`
	Aggregate      minimalAttestation
}

type minimalBeaconState struct {
	GenesisTime       uint64             `json:"genesis_time"`
	Slot              uint64             `json:"slot"`
	Fork              minimalFork        `json:"fork"`
	LatestBlockHeader minimalBlockHeader `json:"latest_block_header"`
	BlockRoots        [][]byte           `json:"block_roots" ssz-size:"64,32"`
	StateRoots        [][]byte           `json:"state_roots" ssz-size:"64,32"`
	HistoricalRoots   [][]byte           `json:"historical_roots" ssz-size:"?,32" ssz-max:"16777216"`
	Eth1Data          minimalEth1Data    `json:"eth1_data"`
	Eth1DataVotes     []minimalEth1Data  `json:"eth1_data_votes" ssz-max:"16"`
	Eth1DepositIndex  uint64             `json:"eth1_deposit_index"`
	Validators        []minimalValidator `json:"validators" ssz-max:"1099511627776"`
	Balances          []uint64           `json:"balances" ssz-max:"1099511627776"`
	RandaoMixes       [][]byte           `json:"randao_mixes" ssz-size:"64,32"`
	Slashings         []uint64           `json:"slashings" ssz-size:"64"`

	PreviousEpochAttestations []minimalPendingAttestation `json:"previous_epoch_attestations" ssz-max:"1024"`
	CurrentEpochAttestations  []minimalPendingAttestation `json:"current_epoch_attestations" ssz-max:"1024"`
	JustificationBits         bitfield.Bitvector4         `json:"justification_bits" ssz-size:"1"`

	PreviousJustifiedCheckpoint minimalCheckpoint `json:"previous_justified_checkpoint"`
	CurrentJustifiedCheckpoint  minimalCheckpoint `json:"current_justified_checkpoint"`
	FinalizedCheckpoint         minimalCheckpoint `json:"finalized_checkpoint"`
}
