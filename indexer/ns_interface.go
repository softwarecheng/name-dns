package indexer

import (
	"github.com/OLProtocol/ordx/common"
	"github.com/OLProtocol/ordx/indexer/ns"
)

func (b *IndexerMgr) GetNSStatus() *common.NameServiceStatus {
	return b.ns.GetStatus()
}

func (b *IndexerMgr) getNameInfoWithRegInfo(reg *ns.NameRegister) *common.NameInfo {
	address := b.GetAddressById(reg.Nft.OwnerAddressId)
	utxo := b.GetUtxoById(reg.Nft.UtxoId)
	kvs := make(map[string]*common.KeyValueInDB)
	attr := b.ns.GetNameProperties(reg.Name)
	if attr != nil {
		for k, v := range attr.KVs {
			kvs[k] = &common.KeyValueInDB{Value: v.Value, InscriptionId: v.InscriptionId}
		}
	}

	return &common.NameInfo{
		Base:         reg.Nft.Base,
		Id:           reg.Id,
		Name:         reg.Name,
		OwnerAddress: address,
		Utxo:         utxo,
		KVs:          kvs,
	}
}

func (b *IndexerMgr) GetNameInfo(name string) *common.NameInfo {
	reg := b.ns.GetNameRegisterInfo(name)
	if reg == nil {
		common.Log.Errorf("GetNameRegisterInfo %s failed", name)
		return nil
	}

	return b.getNameInfoWithRegInfo(reg)
}

func (b *IndexerMgr) IsNameExist(name string) bool {
	return b.ns.IsNameExist(name)
}

func (b *IndexerMgr) GetNamesWithUtxo(utxoId uint64) []string {
	return b.ns.GetNamesWithUtxo2(utxoId)
}

func (b *IndexerMgr) GetNames(start, limit int) []string {
	return b.ns.GetNames(start, limit)
}

func (b *IndexerMgr) GetNamesWithSat(sat int64) []*common.NameInfo {
	result := make([]*common.NameInfo, 0)

	names := b.ns.GetNameRegisterInfoWithSat(sat)
	for _, name := range names {
		info := b.getNameInfoWithRegInfo(name)
		if info != nil {
			result = append(result, info)
		}
	}

	return result
}

func (p *IndexerMgr) GetNameHistory(start int, limit int) []*common.MintAbbrInfo {
	result := make([]*common.MintAbbrInfo, 0)
	names := p.ns.GetNames(start, limit)
	for _, name := range names {
		reg := p.ns.GetNameRegisterInfo(name)
		if reg != nil {
			info := common.NewMintAbbrInfo2(reg.Nft.Base)
			result = append(result, info)
		}
	}
	return result
}

func (p *IndexerMgr) GetNameHistoryWithAddress(addressId uint64, start int, limit int) ([]*common.MintAbbrInfo, int) {
	result := make([]*common.MintAbbrInfo, 0)
	nfts, total := p.ns.GetNamesWithInscriptionAddress(addressId, start, limit)
	for _, nft := range nfts {
		info := common.NewMintAbbrInfo2(nft.Base)
		result = append(result, info)
	}
	return result, total
}
