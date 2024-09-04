package ordx

import (
	"fmt"

	"github.com/OLProtocol/ordx/common"
	serverOrdx "github.com/OLProtocol/ordx/server/define"
)

func (s *Model) GetNSStatusList(start, limit int) (*NSStatusData, error) {
	status := s.indexer.GetNSStatus()

	ret := NSStatusData{Version: status.Version, Total: (status.NameCount), Start: uint64(start)}
	names := s.indexer.GetNames(start, limit)
	for _, name := range names {
		info := s.indexer.GetNameInfo(name)
		if info != nil {
			item := s.nameToItem(info)
			ret.Names = append(ret.Names, item)
		}
	}

	return &ret, nil
}

func (s *Model) GetNameInfo(name string) (*serverOrdx.OrdinalsName, error) {
	info := s.indexer.GetNameInfo(name)
	if info == nil {
		return nil, fmt.Errorf("can't find name %s", name)
	}

	ret := serverOrdx.OrdinalsName{NftItem: *s.nameToItem(info)}
	for k, v := range info.KVs {
		item := serverOrdx.KVItem{Key: k, Value: v.Value, InscriptionId: v.InscriptionId}
		ret.KVItemList = append(ret.KVItemList, &item)
	}

	return &ret, nil
}

func (s *Model) GetNameRouting(name string) (*serverOrdx.NameRouting, error) {
	info := s.indexer.GetNameInfo(name)
	if info == nil {
		return nil, fmt.Errorf("can't find name %s", name)
	}

	ret := serverOrdx.NameRouting{Holder: info.OwnerAddress, InscriptionId: info.Base.InscriptionId, P: "btcname", Op: "routing", Name: info.Name}
	for k, v := range info.KVs {
		switch k {
		case "ord_handle":
			ret.Handle = v.Value
		case "ord_index":
			ret.Index = v.Value
		}
	}

	return &ret, nil
}

func (s *Model) baseContentToNftItem(info *common.InscribeBaseContent) *serverOrdx.NftItem {
	return &serverOrdx.NftItem{
		Id:                 info.Id,
		Name:               info.TypeName,
		Sat:                info.Sat,
		InscriptionId:      info.InscriptionId,
		BlockHeight:        int(info.BlockHeight),
		BlockTime:          info.BlockTime,
		InscriptionAddress: s.indexer.GetAddressById(info.InscriptionAddress)}
}

func (s *Model) nameToItem(info *common.NameInfo) *serverOrdx.NftItem {
	item := s.baseContentToNftItem(info.Base)
	item.Address = info.OwnerAddress
	item.Utxo = info.Utxo
	item.Value = s.getUtxoValue2(info.Utxo)
	item.Id = info.Id
	item.Name = info.Name
	return item
}

func (s *Model) getUtxoValue2(utxo string) int64 {
	_, rngs, err := s.indexer.GetOrdinalsWithUtxo(utxo)
	if err != nil {
		return 0
	}
	return common.GetOrdinalsSize(rngs)
}
