package ordx

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/OLProtocol/ordx/common"
	serverOrdx "github.com/OLProtocol/ordx/server/define"
)

func (s *Model) GetDetailAssetWithRanges(req *RangesReq) (*serverOrdx.AssetDetailInfo, error) {

	var result serverOrdx.AssetDetailInfo
	result.Ranges = req.Ranges
	result.Utxo = ""
	result.Value = common.GetOrdinalsSize(req.Ranges)

	assets := s.indexer.GetAssetsWithRanges(req.Ranges)
	for tickerName, info := range assets {

		var tickinfo serverOrdx.TickerAsset
		tickinfo.Ticker = tickerName
		tickinfo.Utxo = ""
		tickinfo.Amount = 0

		for mintutxo, mintranges := range info {
			asset := serverOrdx.InscriptionAsset{}
			asset.AssetAmount = common.GetOrdinalsSize(mintranges)
			asset.Ranges = mintranges
			asset.InscriptionNum = common.INVALID_INSCRIPTION_NUM
			asset.InscriptionID = mintutxo

			tickinfo.Assets = append(tickinfo.Assets, &asset)
			tickinfo.AssetAmount += asset.AssetAmount
		}

		result.Assets = append(result.Assets, &tickinfo)
	}

	sort.Slice(result.Assets, func(i, j int) bool {
		return result.Assets[i].AssetAmount > result.Assets[j].AssetAmount
	})

	return &result, nil
}

func (s *Model) GetAbbrAssetsWithUtxo(utxo string) ([]*serverOrdx.AssetAbbrInfo, error) {
	result := make([]*serverOrdx.AssetAbbrInfo, 0)
	utxoId := s.indexer.GetUtxoId(utxo)
	assets := s.indexer.GetAssetsWithUtxo(utxoId)
	for ticker, mintinfo := range assets {

		amount := int64(0)
		for _, rng := range mintinfo {
			amount += common.GetOrdinalsSize(rng)
		}

		result = append(result, &serverOrdx.AssetAbbrInfo{
			TypeName: ticker.Name,
			Ticker:   ticker.Name,
			Amount:   amount,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Amount > result[j].Amount
	})

	return result, nil
}

func (s *Model) GetSeedsWithUtxo(utxo string) ([]*serverOrdx.Seed, error) {
	result := make([]*serverOrdx.Seed, 0)
	assets := s.indexer.GetAssetsWithUtxo(s.indexer.GetUtxoId(utxo))
	for ticker, info := range assets {
		assetRanges := make([]*common.Range, 0)
		for _, rngs := range info {
			assetRanges = append(assetRanges, rngs...)
		}
		seed := serverOrdx.Seed{TypeName: ticker.TypeName, Ticker: ticker.Name, Seed: common.GenerateSeed2(assetRanges)}
		result = append(result, &seed)
	}

	return result, nil
}

func (s *Model) GetSatRangeWithUtxo(utxo string) (*serverOrdx.UtxoInfo, error) {
	utxoId := uint64(common.INVALID_ID)
	if len(utxo) < 64 {
		utxoId, _ = strconv.ParseUint(utxo, 10, 64)
	}

	result := serverOrdx.UtxoInfo{}
	var err error
	if utxoId == common.INVALID_ID {
		result.Id, result.Ranges, err = s.indexer.GetOrdinalsWithUtxo(utxo)
		result.Utxo = utxo
	} else {
		result.Utxo, result.Ranges, err = s.indexer.GetOrdinalsWithUtxoId(utxoId)
		result.Id = utxoId
	}
	if err != nil {
		common.Log.Warnf("GetSatRangeWithUtxo %s failed, %v", utxo, err)
		return nil, err
	}

	return &result, nil
}

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

func (s *Model) GetNameValues(name, prefix string, start, limit int) (*serverOrdx.OrdinalsName, error) {
	info := s.indexer.GetNameInfo(name)
	if info == nil {
		return nil, fmt.Errorf("can't find name %s", name)
	}

	type FilterResult struct {
		Key   string
		Value *common.KeyValueInDB
	}

	filter := make([]*FilterResult, 0)
	for k, v := range info.KVs {
		if strings.HasPrefix(k, prefix) {
			filter = append(filter, &FilterResult{Key: k, Value: v})
		}
	}

	sort.Slice(filter, func(i, j int) bool {
		return filter[i].Key > filter[j].Key
	})

	totalRecords := len(filter)
	if totalRecords < start {
		return nil, fmt.Errorf("start exceeds boundary")
	}
	if totalRecords < start+limit {
		limit = totalRecords - start
	}
	end := start + limit
	newFilter := filter[start:end]

	ret := serverOrdx.OrdinalsName{NftItem: *s.nameToItem(info)}
	for _, kv := range newFilter {
		item := serverOrdx.KVItem{Key: kv.Key, Value: kv.Value.Value, InscriptionId: kv.Value.InscriptionId}
		ret.KVItemList = append(ret.KVItemList, &item)
	}
	ret.Total = totalRecords
	ret.Start = start

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

func (s *Model) GetNamesWithAddress(address, sub string, start, limit int) (*NamesWithAddressData, error) {
	ret := NamesWithAddressData{Address: address}
	var names []*common.NameInfo
	var total int
	if sub == "" {
		names, total = s.indexer.GetNamesWithAddress(address, start, limit)
	} else {
		if sub == "PureName" {
			sub = ""
		}
		names, total = s.indexer.GetSubNamesWithAddress(address, sub, start, limit)
	}

	for _, info := range names {
		data := serverOrdx.OrdinalsName{NftItem: *s.nameToItem(info)}
		for k, v := range info.KVs {
			item := serverOrdx.KVItem{Key: k, Value: v.Value, InscriptionId: v.InscriptionId}
			data.KVItemList = append(data.KVItemList, &item)
		}
		ret.Names = append(ret.Names, &data)
	}
	ret.Total = total

	return &ret, nil
}

func (s *Model) GetNamesWithSat(sat int64) (*NamesWithAddressData, error) {
	ret := NamesWithAddressData{}
	names := s.indexer.GetNamesWithSat(sat)
	for _, info := range names {
		data := serverOrdx.OrdinalsName{NftItem: *s.nameToItem(info)}
		for k, v := range info.KVs {
			item := serverOrdx.KVItem{Key: k, Value: v.Value, InscriptionId: v.InscriptionId}
			data.KVItemList = append(data.KVItemList, &item)
		}
		ret.Names = append(ret.Names, &data)
	}
	ret.Total = len(names)

	sort.Slice(ret.Names, func(i, j int) bool {
		return ret.Names[i].Name < ret.Names[j].Name
	})

	return &ret, nil
}

func (s *Model) GetNameWithInscriptionId(id string) (*serverOrdx.OrdinalsName, error) {
	info := s.indexer.GetNameWithInscriptionId(id)
	if info == nil {
		return nil, fmt.Errorf("can't find name with %s", id)
	}

	ret := serverOrdx.OrdinalsName{NftItem: *s.nameToItem(info)}
	for k, v := range info.KVs {
		item := serverOrdx.KVItem{Key: k, Value: v.Value, InscriptionId: v.InscriptionId}
		ret.KVItemList = append(ret.KVItemList, &item)
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
