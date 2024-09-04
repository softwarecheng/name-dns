package indexer

import (
	"github.com/OLProtocol/ordx/common"
	"github.com/btcsuite/btcd/chaincfg"
)

// interface for RPC

func (b *IndexerMgr) IsMainnet() bool {
	return b.chaincfgParam.Name == "mainnet"
}

func (b *IndexerMgr) GetBaseDBVer() string {
	return b.compiling.GetBaseDBVer()
}

func (b *IndexerMgr) GetChainParam() *chaincfg.Params {
	return b.chaincfgParam
}

// return: addressId -> asset amount

func (b *IndexerMgr) GetNftWithInscriptionId(inscriptionId string) *common.Nft {
	return b.nft.GetNftWithInscriptionId(inscriptionId)
}
