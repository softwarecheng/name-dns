package define

import (
	"github.com/OLProtocol/ordx/common"
)

type TickerStatus struct {
	ID              int64  `json:"id" example:"1"`
	Ticker          string `json:"ticker" example:"BTC"`
	StartBlock      int    `json:"startBlock" example:"100"`
	EndBlock        int    `json:"endBlock" example:"200"`
	TotalMinted     int64  `json:"totalMinted" example:"546"`
	Limit           int64  `json:"limit" example:"100"`
	SelfMint        int    `json:"selfmint" example:"100"`
	Max             int64  `json:"max" example:"10000"`
	DeployHeight    int    `json:"deployHeight" example:"100"`
	MintTimes       int64  `json:"mintTimes" example:"100"`
	HoldersCount    int    `json:"holdersCount" example:"100"`
	DeployBlocktime int64  `json:"deployBlocktime" example:"100"`
	InscriptionId   string `json:"inscriptionId" example:"bac89275b4c0a0ba6aaa603d749a1c88ae3033da9f6d6e661a28fb40e8dca362i0"`
	InscriptionNum  int64  `json:"inscriptionNum" example:"67269474"`
	Description     string `json:"description" example:"xxx"`
	Rarity          string `json:"rarity" example:"xxx"`
	DeployAddress   string `json:"deployAddress" example:"bc1p9jh2caef2ejxnnh342s4eaddwzntqvxsc2cdrsa25pxykvkmgm2sy5ycc5"`
	Content         []byte `json:"content,omitempty"`
	ContentType     string `json:"contenttype,omitempty" example:"xxx"`
	Delegate        string `json:"delegate,omitempty" example:"xxx"`
	TxId            string `json:"txid" example:"xxx"`
}

type MintDetailInfo struct {
	ID             int64           `json:"id" example:"1"`
	Ticker         string          `json:"ticker,omitempty"`
	MintAddress    string          `json:"address,omitempty"`
	Amount         int64           `json:"amount,omitempty"`
	MintTime       int64           `json:"mintTimes,omitempty"`
	Delegate       string          `json:"delegate,omitempty"`
	Content        []byte          `json:"content,omitempty"`
	ContentType    string          `json:"contenttype,omitempty"`
	Ranges         []*common.Range `json:"ranges,omitempty"`
	InscriptionID  string          `json:"inscriptionId,omitempty"`
	InscriptionNum int64           `json:"inscriptionNumber,omitempty"`
}

type MintPermissionInfo struct {
	Ticker  string `json:"ticker"`
	Address string `json:"address"`
	Amount  int64  `json:"amount"`
}

type FeeInfo struct {
	Address  string `json:"address"`
	Discount int    `json:"discount"` // 0-100: 100 means free mint
}

type MintHistoryItem struct {
	MintAddress    string `json:"mintaddress,omitempty" example:"bc1p9jh2caef2ejxnnh342s4eaddwzntqvxsc2cdrsa25pxykvkmgm2sy5ycc5"`
	HolderAddress  string `json:"holderaddress,omitempty"`
	Balance        int64  `json:"balance,omitempty" example:"546" description:"Balance of the holder"`
	InscriptionID  string `json:"inscriptionId,omitempty" example:"bac89275b4c0a0ba6aaa603d749a1c88ae3033da9f6d6e661a28fb40e8dca362i0"`
	InscriptionNum int64  `json:"inscriptionNumber,omitempty" example:"67269474" description:"Inscription number of the holder"`
}

type MintHistory struct {
	TypeName string             `json:"type"`
	Ticker   string             `json:"ticker,omitempty"`
	Total    int                `json:"total,omitempty"`
	Start    int                `json:"start,omitempty"`
	Limit    int                `json:"limit,omitempty"`
	Items    []*MintHistoryItem `json:"items,omitempty"`
}

type Holder struct {
	Wallet       string `json:"wallet,omitempty"`
	TotalBalance int64  `json:"total_balance,omitempty"`
}

type BalanceSummary struct {
	TypeName string `json:"type"`
	Ticker   string `json:"ticker"`
	Balance  int64  `json:"balance"`
}

type InscriptionAsset struct {
	TypeName       string          `json:"type,omitempty"`
	Ticker         string          `json:"ticker,omitempty"`
	InscriptionID  string          `json:"inscriptionId,omitempty"`
	InscriptionNum int64           `json:"inscriptionnum,omitempty"`
	AssetAmount    int64           `json:"assetamount,omitempty"`
	Ranges         []*common.Range `json:"ranges,omitempty"`
}

type TickerAsset struct {
	TypeName    string              `json:"type,omitempty"`
	Ticker      string              `json:"ticker,omitempty"`
	Utxo        string              `json:"utxo,omitempty"`
	Amount      int64               `json:"amount,omitempty"`
	AssetAmount int64               `json:"assetamount,omitempty"`
	Assets      []*InscriptionAsset `json:"assets,omitempty"`
}

type AssetDetailInfo struct {
	Utxo   string          `json:"utxo,omitempty"`
	Value  int64           `json:"value,omitempty"`
	Ranges []*common.Range `json:"ranges,omitempty"`
	Assets []*TickerAsset  `json:"assets,omitempty"`
}

type UtxoSort struct {
	Utxo string
	Ts   int64
}

type AssetAbbrInfo struct {
	TypeName string `json:"type"`
	Ticker   string `json:"ticker"`
	Amount   int64  `json:"amount"`
}

type UtxoAbbrAssets struct {
	Utxo   string             `json:"utxo"`
	Assets []*AssetAbbrInfo   `json:"assets"`
}

type Seed struct {
	TypeName string `json:"type"`
	Ticker   string `json:"ticker"`
	Seed     string `json:"seed"`
}

type UtxoInfo struct {
	Utxo   string          `json:"utxo"`
	Id     uint64          `json:"id"`
	Ranges []*common.Range `json:"ranges,omitempty"`
}

type NftItem struct {
	Id                 int64  `json:"id"`
	Name               string `json:"name"`
	Sat                int64  `json:"sat"`
	Address            string `json:"address"`
	InscriptionId      string `json:"inscriptionId"`
	Utxo               string `json:"utxo"`
	Value              int64  `json:"value"`
	BlockHeight        int    `json:"height"`
	BlockTime          int64  `json:"time"`
	InscriptionAddress string `json:"inscriptionAddress"`
}

type KVItem struct {
	Key           string `json:"key"`
	Value         string `json:"value"`
	InscriptionId string `json:"inscriptionId"`
}

type OrdinalsName struct {
	NftItem
	Total      int       `json:"total,omitempty"`
	Start      int       `json:"start,omitempty"`
	KVItemList []*KVItem `json:"kvs"`
}

type NameRouting struct {
	Holder           string  `json:"holder"`
	InscriptionId    string  `json:"inscription_id"`
	P                string  `json:"p"`
	Op               string  `json:"op"`
	Name             string  `json:"name"`
	Handle           string  `json:"ord_handle"`
	Index            string  `json:"ord_index"`
}

type NameCheckResult struct {
	Name   string `json:"name"`
	Result int    `json:"result"` // 0 允许铸造； 1 已经铸造； < 0，其他错误
}

type InscriptionId struct {
	Id   string `json:"id"`
}

type NftInfo struct {
	NftItem
	ContentType  []byte `json:"contenttype"`
	Content      []byte `json:"content"`
	MetaProtocol []byte `json:"metaprotocol"`
	MetaData     []byte `json:"metadata"`
	Parent       string `json:"parent"`
	Delegate     string `json:"delegate"`
}
