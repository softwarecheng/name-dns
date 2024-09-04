package base_indexer

import (
	"github.com/OLProtocol/ordx/common"
	"github.com/btcsuite/btcd/chaincfg"
)

type Indexer interface {
	IsMainnet() bool
	GetChainParam() *chaincfg.Params
	GetBaseDBVer() string
	GetOrdxDBVer() string
	GetChainTip() int
	GetSyncHeight() int
	GetBlockInfo(int) (*common.BlockInfo, error)

	// base indexer
	GetAddressById(addressId uint64) string
	GetAddressId(address string) uint64
	GetUtxoById(utxoId uint64) string
	GetUtxoId(utxo string) uint64
	// return: utxoId->value
	GetUTXOsWithAddress(address string) (map[uint64]int64, error)
	// return: utxo, sat ranges
	GetOrdinalsWithUtxoId(id uint64) (string, []*common.Range, error)
	// return: utxoId, sat ranges
	GetOrdinalsWithUtxo(utxo string) (uint64, []*common.Range, error)
	// return: address, utxo

	// FT

	// NameService
	GetNSStatus() *common.NameServiceStatus
	IsNameExist(name string) bool
	GetNameInfo(name string) *common.NameInfo
	GetNamesWithUtxo(utxoId uint64) []string
	GetNames(start, limit int) []string
	GetNamesWithSat(sat int64) []*common.NameInfo

	// ntf
	GetNftStatus() *common.NftStatus
	GetNftInfo(id int64) *common.Nft
	GetNftInfoWithInscriptionId(inscriptionId string) *common.Nft
	GetNftsWithUtxo(utxoId uint64) []string
	GetNftsWithSat(sat int64) *common.NftsInSat
	GetNfts(start, limit int) ([]int64, int)
	GetNftsWithAddress(address string, start int, limit int) ([]*common.Nft, int)
}
