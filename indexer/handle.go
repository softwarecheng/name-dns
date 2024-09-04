package indexer

import (
	"strconv"
	"strings"
	"time"

	"github.com/OLProtocol/ordx/common"
	"github.com/OLProtocol/ordx/indexer/ns"
)

func (s *IndexerMgr) processOrdProtocol(block *common.Block) {
	if block.Height < s.ordFirstHeight {
		return
	}

	detectOrdMap := make(map[string]int, 0)
	measureStartTime := time.Now()
	//common.Log.Info("processOrdProtocol ...")
	count := 0
	for _, tx := range block.Transactions {
		id := 0
		for _, input := range tx.Inputs {

			inscriptions, err := common.ParseInscription(input.Witness)
			if err != nil {
				continue
			}

			for _, insc := range inscriptions {
				s.handleOrd(input, insc, id, tx, block)
				id++
				count++
			}
		}
		if id > 0 {
			detectOrdMap[tx.Txid] = id
		}
	}
	//common.Log.Infof("processOrdProtocol loop %d finished. cost: %v", count, time.Since(measureStartTime))

	//time2 := time.Now()
	s.exotic.UpdateTransfer(block)
	s.nft.UpdateTransfer(block)
	s.ns.UpdateTransfer(block)
	s.ftIndexer.UpdateTransfer(block)

	//common.Log.Infof("processOrdProtocol UpdateTransfer finished. cost: %v", time.Since(time2))

	// 检测是否一致，如果不一致，需要进一步调试。
	// s.detectInconsistent(detectOrdMap, block.Height)

	common.Log.Infof("processOrdProtocol %d,is done: cost: %v", block.Height, time.Since(measureStartTime))
}

func findOutputWithSat(tx *common.Transaction, sat int64) *common.Output {
	for _, out := range tx.Outputs {
		for _, rng := range out.Ordinals {
			if sat >= rng.Start && sat < rng.Start+rng.Size {
				return out
			}
		}
	}
	return nil
}

func (s *IndexerMgr) handleNameRegister(content *common.OrdxRegContent, nft *common.Nft) {

	name := strings.ToLower(content.Name)

	reg := &ns.NameRegister{
		Nft:  nft,
		Name: name,
	}
	nft.Base.TypeName = common.ASSET_TYPE_NS
	nft.Base.UserData = []byte(name)

	s.ns.NameRegister(reg)

	if len(content.KVs) > 0 {
		update := &ns.NameUpdate{
			InscriptionId: nft.Base.InscriptionId,
			BlockHeight:   int(nft.Base.BlockHeight),
			Sat:           nft.Base.Sat,
			Name:          name,
			KVs:           ns.ParseKVs(content.KVs),
		}
		s.ns.NameUpdate(update)
	}
}

func (s *IndexerMgr) handleNameUpdate(content *common.OrdxUpdateContentV2, nft *common.Nft) {

	content.Name = strings.ToLower(content.Name)

	reg := s.ns.GetNameRegisterInfo(content.Name)
	if reg == nil {
		common.Log.Warnf("IndexerMgr.handleNameUpdate: %s, Name %s not exist", nft.Base.InscriptionId, content.Name)
		return
	}

	// 只需要当前owner持有该nft就可以修改，而不必在sat上继续铸造
	if nft.OwnerAddressId != reg.Nft.OwnerAddressId {
		common.Log.Warnf("IndexerMgr.handleNameUpdate: %s, Name %s has different owner", nft.Base.InscriptionId, content.Name)
		return
	}

	kvs := make([]*ns.KeyValue, 0)
	for k, v := range content.KVs {
		// 对于需要做持有者检查的属性，简单忽略就行，不影响其他有效属性
		if k == "avatar" {
			avatar := s.nft.GetNftWithInscriptionId(v)
			if avatar == nil || avatar.OwnerAddressId != nft.OwnerAddressId {
				common.Log.Warnf("IndexerMgr.handleNameUpdate: %s, name: %s, invalid avatar: %v, ignore it",
					nft.Base.InscriptionId, content.Name, v)
				continue
			}
		}
		kvs = append(kvs, &ns.KeyValue{Key: k, Value: v})
	}

	update := &ns.NameUpdate{
		InscriptionId: nft.Base.InscriptionId,
		BlockHeight:   int(nft.Base.BlockHeight),
		Name:          content.Name,
		KVs:           kvs,
	}
	nft.Base.TypeName = common.ASSET_TYPE_NFT

	s.ns.NameUpdate(update)
}

func (s *IndexerMgr) handleNameRouting(content *common.OrdxUpdateContentV2, nft *common.Nft) {

	content.Name = strings.ToLower(content.Name)

	reg := s.ns.GetNameRegisterInfo(content.Name)
	if reg == nil {
		common.Log.Warnf("IndexerMgr.handleNameRouting: %s, Name %s not exist", nft.Base.InscriptionId, content.Name)
		return
	}

	// 只需要当前owner持有该nft就可以修改，而不必在sat上继续铸造
	if nft.OwnerAddressId != reg.Nft.OwnerAddressId {
		common.Log.Warnf("IndexerMgr.handleNameRouting: %s, Name %s has different owner", nft.Base.InscriptionId, content.Name)
		return
	}

	kvs := make([]*ns.KeyValue, 0)
	for k, v := range content.KVs {
		kvs = append(kvs, &ns.KeyValue{Key: k, Value: v})
	}

	update := &ns.NameUpdate{
		InscriptionId: nft.Base.InscriptionId,
		BlockHeight:   int(nft.Base.BlockHeight),
		Name:          content.Name,
		KVs:           kvs,
	}
	nft.Base.TypeName = common.ASSET_TYPE_NFT

	s.ns.NameUpdate(update)
}

func (s *IndexerMgr) handleOrd(input *common.Input,
	fields map[int][]byte, inscriptionId int, tx *common.Transaction, block *common.Block) {

	satpoint := 0
	if fields[common.FIELD_POINT] != nil {
		satpoint = common.GetSatpoint(fields[common.FIELD_POINT])
		if int64(satpoint) >= common.GetOrdinalsSize(input.Ordinals) {
			satpoint = 0
		}
	}

	var output *common.Output
	sat := getSatInRange(input.Ordinals, satpoint)
	if sat > 0 {
		output = findOutputWithSat(tx, sat)
		if output == nil {
			output = findOutputWithSat(block.Transactions[0], sat)
			if output == nil {
				common.Log.Panicf("processOrdProtocol: tx: %s, findOutputWithSat %d failed", tx.Txid, sat)
			}
		}
	} else {
		output = tx.Outputs[0]
	}

	// 1. 先保存nft数据
	nft := s.handleNft(input, output, satpoint, fields, inscriptionId, tx, block)
	if nft == nil {
		return
	}

	if len(input.Ordinals) == 0 {
		// 虽然ordinals.com解析出了这个交易，但是我们认为该交易没有输入的sat，也就是无法将数据绑定到某一个sat上，违背了协议原则
		// 特殊交易，ordx不支持，不处理
		// c1e0db6368a43f5589352ed44aa1ff9af33410e4a9fd9be0f6ac42d9e4117151
		// TODO 0605版本中，没有把这个nft编译进来
		return
	}

	// 2. 再看看是否ordx协议
	protocol, content := common.GetProtocol(fields)
	switch protocol {
	case "sns":
		domain := common.ParseDomainContent(string(fields[common.FIELD_CONTENT]))
		if domain == nil {
			domain = common.ParseDomainContent(string(content))
		}
		if domain != nil {
			switch domain.Op {
			case "reg":
				s.handleSnsName(domain.Name, nft)
			case "update":
				var updateInfo *common.OrdxUpdateContentV2
				// 如果有metadata，那么不处理FIELD_CONTENT的内容
				if string(fields[common.FIELD_META_PROTOCOL]) == "sns" && fields[common.FIELD_META_DATA] != nil {
					updateInfo = common.ParseUpdateContent(string(content))
					updateInfo.P = "sns"
					value, ok := updateInfo.KVs["key"]
					if ok {
						delete(updateInfo.KVs, "key")
						updateInfo.KVs[value] = nft.Base.InscriptionId
					}
				} else {
					updateInfo = common.ParseUpdateContent(string(fields[common.FIELD_CONTENT]))
				}

				if updateInfo == nil {
					return
				}
				s.handleNameUpdate(updateInfo, nft)
			}
		}
	case "brc-20":
		brc20Content := common.ParseBrc20Content(string(fields[common.FIELD_CONTENT]))
		if brc20Content != nil {
			switch brc20Content.Op {
			case "deploy":
				s.handleSnsName(brc20Content.Ticker, nft)
			}
		}
	case "primary-name":
		primaryNameContent := common.ParseCommonContent(string(fields[common.FIELD_CONTENT]))
		if primaryNameContent != nil {
			switch primaryNameContent.Op {
			case "update":
				s.handleNameUpdate(primaryNameContent, nft)
			}
		}
		// {
		// 	"p": "primary-name",
		// 	"op": "update",
		// 	"name": "btcname.btc",
		// 	"avatar": "41479dbcb749ec04872b77c5cb4a67dc7b13f746ba2e86ba70854d0cdaed0646i0"
		//   }
		// type: application/json
		// content: { "p": "sns", "op": "reg", "name": "1866.sats"}
		// or ： text/plain;charset=utf-8 {"p":"sns","op":"reg","name":"good.sats"}
	case "btcname":
		commonContent := common.ParseCommonContent(string(fields[common.FIELD_CONTENT]))
		if commonContent != nil {
			switch commonContent.Op {
			case "routing":
				s.handleNameRouting(commonContent, nft)
			}
		}
		/*
			{
				"p":"btcname",
				"op":"routing",
				"name":"xxx.btc",
				"ord_handle":"xxx",
				"ord_index":"xxxi0",
				"btc_p2phk":"1xxx",
				"btc_p2sh":"3xxx",
				"btc_segwit":"bc1qxxx",
				"btc_lightning":"xxx",
				"eth_address":"0xxxx",
				"matic_address":"0xxxx",
				"sol_address":"xxx",
				"avatar":"xxxi0"
			}
		*/

	default:
		// 3. 如果content中的内容格式，符合 *.* 或者 * , 并且字段在32字节以内，符合名字规范，就把它当做一个名字来处理
		// text/plain;charset=utf-8 abc
		// 或者简单文本 xxx.xx 或者 xx
		if protocol == "" {
			s.handleSnsName(string(fields[common.FIELD_CONTENT]), nft)
		}
	}

}

func (s *IndexerMgr) handleSnsName(name string, nft *common.Nft) {
	name = common.PreprocessName(name)
	if common.IsValidSNSName(name) {
		info := s.ns.GetNameRegisterInfo(name)
		if info != nil {
			common.Log.Warnf("%s Name %s exist, registered at %s",
				nft.Base.InscriptionId, name, info.Nft.Base.InscriptionId)
			return
		}

		regInfo := &common.OrdxRegContent{
			OrdxBaseContent: common.OrdxBaseContent{P: "sns", Op: "reg"},
			Name:            name}

		s.handleNameRegister(regInfo, nft)
	}
}

func (s *IndexerMgr) handleNft(input *common.Input, output *common.Output, satpoint int,
	fields map[int][]byte, inscriptionId int, tx *common.Transaction, block *common.Block) *common.Nft {

	//if s.nft.Base.IsEnabled() {
	sat := int64(-1)
	if len(input.Ordinals) > 0 {
		newRngs := reAlignRange(input.Ordinals, satpoint, 1)
		sat = newRngs[0].Start
	}

	//addressId1 := s.compiling.GetAddressId(input.Address.Addresses[0])
	addressId2 := s.compiling.GetAddressId(output.Address.Addresses[0])
	utxoId := common.GetUtxoId(output)
	nft := common.Nft{
		Base: &common.InscribeBaseContent{
			InscriptionId:      tx.Txid + "i" + strconv.Itoa(inscriptionId),
			InscriptionAddress: addressId2, // TODO 这个地址不是铭刻者，模型的问题，比较难改，直接使用输出地址
			BlockHeight:        int32(block.Height),
			BlockTime:          block.Timestamp.Unix(),
			ContentType:        (fields[common.FIELD_CONTENT_TYPE]),
			Content:            fields[common.FIELD_CONTENT],
			ContentEncoding:    fields[common.FIELD_CONTENT_ENCODING],
			MetaProtocol:       (fields[common.FIELD_META_PROTOCOL]),
			MetaData:           fields[common.FIELD_META_DATA],
			Parent:             common.ParseInscriptionId(fields[common.FIELD_PARENT]),
			Delegate:           common.ParseInscriptionId(fields[common.FIELD_DELEGATE]),
			Sat:                sat,
			TypeName:           common.ASSET_TYPE_NFT,
		},
		OwnerAddressId: addressId2,
		UtxoId:         utxoId,
	}
	s.nft.NftMint(&nft)
	return &nft
	// }
	// return nil
}

func getSatInRange(common []*common.Range, satpoint int) int64 {
	for _, rng := range common {
		if satpoint >= int(rng.Size) {
			satpoint -= int(rng.Size)
		} else {
			return rng.Start + int64(satpoint)
		}
	}

	return -1
}
