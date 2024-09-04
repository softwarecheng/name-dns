package common

import (
	"math"

	"github.com/OLProtocol/ordx/common/pb"
)

const (
	DB_KEY_UTXO         = "u-"  // utxo -> UtxoValueInDB
	DB_KEY_ADDRESS      = "a-"  // address -> addressId
	DB_KEY_ADDRESSVALUE = "av-" // addressId-utxoId -> value
	DB_KEY_UTXOID       = "ui-" // utxoId -> utxo
	DB_KEY_ADDRESSID    = "ai-" // addressId -> address
	DB_KEY_BLOCK        = "b-"  // height -> block
)

// Address Type defined in txscript.ScriptClass

type UtxoValueInDB = pb.MyUtxoValueInDB

type UtxoIdInDB struct {
	UtxoId uint64
	Value  int64
}

type UtxoValue struct {
	Op    int // -1 deleted; 0 read from db; 1 added
	Value int64
}

type AddressValueInDB struct {
	AddressType uint32
	AddressId uint64
	Op        int                   // -1 deleted; 0 read from db; 1 added
	Utxos     map[uint64]*UtxoValue // utxoid -> value
}

type AddressValue struct {
	AddressType uint32
	AddressId uint64
	Utxos     map[uint64]int64 // utxoid -> value
}

type BlockValueInDB struct {
	Height     int
	Timestamp  int64
	InputUtxo  int
	OutputUtxo int
	InputSats  int64
	OutputSats int64
	Ordinals   Range
	LostSats   []*Range // ordinals protocol issue
}

type BlockInfo struct {
	Height     int      `json:"height"`
	Timestamp  int64    `json:"timestamp"`
	TotalSats  int64    `json:"totalsats"`
	RewardSats int64    `json:"rewardsats"`
}

const INVALID_ID = math.MaxUint64
