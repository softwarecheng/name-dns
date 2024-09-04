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
	FindSat(sat int64) (string, string, error)
	// return: address
	GetHolderAddress(inscriptionId string) string

	// Asset
	// return: tick->amount
	GetAssetSummaryInAddress(address string) map[common.TickerName]int64
	// return: tick->UTXOs
	GetAssetUTXOsInAddress(address string) map[*common.TickerName][]uint64
	// return: utxo->asset amount
	GetAssetUTXOsInAddressWithTick(address string, tickerName *common.TickerName) (map[uint64]int64, error)
	// return: mint info sorted by inscribed time
	GetMintHistoryWithAddress(address string, tickerName *common.TickerName, start, limit int) ([]*common.MintAbbrInfo, int)
	HasAssetInUtxo(utxo string, excludingExotic bool) bool
	// return: ticker -> asset info (inscriptinId -> asset ranges)
	GetAssetsWithUtxo(utxo uint64) map[*common.TickerName]map[string][]*common.Range
	// return: ticker -> asset info (inscriptinId -> asset ranges)
	GetAssetsWithRanges([]*common.Range) map[string]map[string][]*common.Range
	// return: exotic range and types
	GetExoticsWithRanges(ranges []*common.Range) []*common.ExoticRange
	HasExoticInRanges(ranges []*common.Range) bool
	GetExoticsWithType(ranges []*common.Range, typ string) []*common.ExoticRange
	AddCollection(ntype, ticker string, ids []string) error
	
	// FT
	// return: ticker's name -> ticker info
	GetTickerMap() (map[string]*common.Ticker, error)
	// return: ticker info
	GetTicker(tickerName string) *common.Ticker
	// return: addressId -> asset amount
	GetHoldersWithTick(tickerName string) map[uint64]int64
	// return: holder amount
	GetHolderAmountWithTick(tickerName string) int
	// return: asset amount, mint times
	GetMintAmount(tickerName string) (int64, int64)
	// return: mint info
	GetMintInfo(inscription string) *common.Mint
	// return: permitted amount to mint
	GetMintPermissionInfo(ticker, address string) int64
	// return:  mint info sorted by inscribed time
	GetMintHistory(tickerName string, start, limit int) []*common.MintAbbrInfo
	// return: inscriptionIds that are splitted.
	GetSplittedInscriptionsWithTick(tickerName string) []string

	
	// NameService
	GetNSStatus() *common.NameServiceStatus
	IsNameExist(name string) bool
	GetNameInfo(name string) *common.NameInfo
	GetNameWithInscriptionId(id string) *common.NameInfo
	GetNamesWithUtxo(utxoId uint64) []string
	GetNames(start, limit int) []string
	GetSubNameSummaryWithAddress(address string) map[string]int64
	GetSubNamesWithAddress(address, sub string, start, limit int) ([]*common.NameInfo, int)
	GetNamesWithAddress(address string, start, limit int) ([]*common.NameInfo, int)
	GetNameAmountWithAddress(address string) int
	GetNamesWithSat(sat int64) []*common.NameInfo

	// ntf
	GetNftStatus() *common.NftStatus
	GetNftInfo(id int64) *common.Nft
	GetNftInfoWithInscriptionId(inscriptionId string) *common.Nft
	GetNftsWithUtxo(utxoId uint64) []string
	GetNftsWithSat(sat int64) *common.NftsInSat
	GetNfts(start, limit int) ([]int64, int)
	GetNftsWithAddress(address string, start int, limit int) ([]*common.Nft, int)
	GetNftHistory(start int, limit int) ([]*common.MintAbbrInfo, int)
	GetNftHistoryWithAddress(addressId uint64, start int, limit int) ([]*common.MintAbbrInfo, int)
}
