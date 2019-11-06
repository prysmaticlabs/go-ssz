package spectests

import (
	"github.com/prysmaticlabs/go-bitfield"
)

type mainnetFork struct {
	PreviousVersion []byte `json:"previous_version" ssz-size:"4"`
	CurrentVersion  []byte `json:"current_version" ssz-size:"4"`
	Epoch           uint64 `json:"epoch"`
}

type mainnetCheckpoint struct {
	Epoch uint64 `json:"epoch"`
	Root  []byte `json:"root" ssz-size:"32"`
}

type mainnetValidator struct {
	Pubkey                     []byte `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials      []byte `json:"withdrawal_credentials" ssz-size:"32"`
	EffectiveBalance           uint64 `json:"effective_balance"`
	Slashed                    bool   `json:"slashed"`
	ActivationEligibilityEpoch uint64 `json:"activation_eligibility_epoch"`
	ActivationEpoch            uint64 `json:"activation_epoch"`
	ExitEpoch                  uint64 `json:"exit_epoch"`
	WithdrawableEpoch          uint64 `json:"withdrawable_epoch"`
}

type mainnetAttestationData struct {
	Slot            uint64
	Index           uint64
	BeaconBlockRoot []byte            `json:"beacon_block_root" ssz-size:"32"`
	Source          mainnetCheckpoint `json:"source"`
	Target          mainnetCheckpoint `json:"target"`
}

type mainnetAttestationAndCustodyBit struct {
	Data       mainnetAttestationData `json:"data"`
	CustodyBit bool                   `json:"custody_bit"`
}

type mainnetIndexedAttestation struct {
	CustodyBit0Indices []uint64               `json:"custody_bit_0_indices" ssz-max:"2048"`
	CustodyBit1Indices []uint64               `json:"custody_bit_1_indices" ssz-max:"2048"`
	Data               mainnetAttestationData `json:"data"`
	Signature          []byte                 `json:"signature" ssz-size:"96"`
}

type mainnetPendingAttestation struct {
	AggregationBits bitfield.Bitlist       `json:"aggregation_bits" ssz-max:"2048"`
	Data            mainnetAttestationData `json:"data"`
	InclusionDelay  uint64                 `json:"inclusion_delay"`
	ProposerIndex   uint64                 `json:"proposer_index"`
}

type mainnetEth1Data struct {
	DepositRoot  []byte `json:"deposit_root" ssz-size:"32"`
	DepositCount uint64 `json:"deposit_count"`
	BlockHash    []byte `json:"block_hash" ssz-size:"32"`
}

type mainnetHistoricalBatch struct {
	BlockRoots [][]byte `json:"block_roots" ssz-size:"8192,32"`
	StateRoots [][]byte `json:"state_roots" ssz-size:"8192,32"`
}

type mainnetDepositData struct {
	Pubkey                []byte `json:"pubkey" ssz-size:"48"`
	WithdrawalCredentials []byte `json:"withdrawal_credentials" ssz-size:"32"`
	Amount                uint64 `json:"amount"`
	Signature             []byte `json:"signature" ssz-size:"96"`
}

// MainnetBlockHeader --
type MainnetBlockHeader struct {
	Slot       uint64 `json:"slot"`
	ParentRoot []byte `json:"parent_root" ssz-size:"32"`
	StateRoot  []byte `json:"state_root" ssz-size:"32"`
	BodyRoot   []byte `json:"body_root" ssz-size:"32"`
	Signature  []byte `json:"signature" ssz-size:"96"`
}

type mainnetProposerSlashing struct {
	ProposerIndex uint64             `json:"proposer_index"`
	Header1       MainnetBlockHeader `json:"header_1"`
	Header2       MainnetBlockHeader `json:"header_2"`
}

type mainnetAttesterSlashing struct {
	Attestation1 mainnetIndexedAttestation `json:"attestation_1"`
	Attestation2 mainnetIndexedAttestation `json:"attestation_2"`
}

type mainnetAttestation struct {
	AggregationBits bitfield.Bitlist       `json:"aggregation_bits" ssz-max:"2048"`
	Data            mainnetAttestationData `json:"data"`
	CustodyBits     bitfield.Bitlist       `json:"custody_bits" ssz-max:"2048"`
	Signature       []byte                 `json:"signature" ssz-size:"96"`
}

type mainnetDeposit struct {
	Proof [][]byte           `json:"proof" ssz-size:"33,32"`
	Data  mainnetDepositData `json:"data"`
}

type mainnetVoluntaryExit struct {
	Epoch          uint64 `json:"epoch"`
	ValidatorIndex uint64 `json:"validator_index"`
	Signature      []byte `json:"signature" ssz-size:"96"`
}

type mainnetBlockBody struct {
	RandaoReveal      []byte                    `json:"randao_reveal" ssz-size:"96"`
	Eth1Data          mainnetEth1Data           `json:"eth1_data"`
	Graffiti          []byte                    `json:"graffiti" ssz-size:"32"`
	ProposerSlashings []mainnetProposerSlashing `json:"proposer_slashings" ssz-max:"16"`
	AttesterSlashings []mainnetAttesterSlashing `json:"attester_slashings" ssz-max:"1"`
	Attestations      []mainnetAttestation      `json:"attestations" ssz-max:"128"`
	Deposits          []mainnetDeposit          `json:"deposits" ssz-max:"16"`
	VoluntaryExits    []mainnetVoluntaryExit    `json:"voluntary_exits" ssz-max:"16"`
}

type mainnetBlock struct {
	Slot       uint64           `json:"slot"`
	ParentRoot []byte           `json:"parent_root" ssz-size:"32"`
	StateRoot  []byte           `json:"state_root" ssz-size:"32"`
	Body       mainnetBlockBody `json:"body"`
	Signature  []byte           `json:"signature" ssz-size:"96"`
}

type mainnetAggregateAndProof struct {
	Index          uint64
	SelectionProof []byte `ssz-size:"96"`
	Aggregate      mainnetAttestation
}

type mainnetBeaconState struct {
	GenesisTime       uint64             `json:"genesis_time"`
	Slot              uint64             `json:"slot"`
	Fork              mainnetFork        `json:"fork"`
	LatestBlockHeader MainnetBlockHeader `json:"latest_block_header"`
	BlockRoots        [][]byte           `json:"block_roots" ssz-size:"8192,32"`
	StateRoots        [][]byte           `json:"state_roots" ssz-size:"8192,32"`
	HistoricalRoots   [][]byte           `json:"historical_roots" ssz-size:"?,32" ssz-max:"16777216"`
	Eth1Data          mainnetEth1Data    `json:"eth1_data"`
	Eth1DataVotes     []mainnetEth1Data  `json:"eth1_data_votes" ssz-max:"1024"`
	Eth1DepositIndex  uint64             `json:"eth1_deposit_index"`
	Validators        []mainnetValidator `json:"validators" ssz-max:"1099511627776"`
	Balances          []uint64           `json:"balances" ssz-max:"1099511627776"`
	RandaoMixes       [][]byte           `json:"randao_mixes" ssz-size:"65536,32"`
	Slashings         []uint64           `json:"slashings" ssz-size:"8192"`

	PreviousEpochAttestations []mainnetPendingAttestation `json:"previous_epoch_attestations" ssz-max:"4096"`
	CurrentEpochAttestations  []mainnetPendingAttestation `json:"current_epoch_attestations" ssz-max:"4096"`
	JustificationBits         bitfield.Bitvector4         `json:"justification_bits" ssz-size:"1"`

	PreviousJustifiedCheckpoint mainnetCheckpoint `json:"previous_justified_checkpoint"`
	CurrentJustifiedCheckpoint  mainnetCheckpoint `json:"current_justified_checkpoint"`
	FinalizedCheckpoint         mainnetCheckpoint `json:"finalized_checkpoint"`
}
