package indexer

import (
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
	count := 0
	for _, tx := range block.Transactions {
		id := 0
		for _, input := range tx.Inputs {

			inscriptions, err := common.ParseInscription(input.Witness)
			if err != nil {
				continue
			}

			for _, insc := range inscriptions {
				s.handleOrd(insc)
				id++
				count++
			}
		}
		if id > 0 {
			detectOrdMap[tx.Txid] = id
		}
	}
	common.Log.Infof("processOrdProtocol %d,is done: cost: %v", block.Height, time.Since(measureStartTime))
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
}

func (s *IndexerMgr) handleNameRouting(content *common.OrdxUpdateContentV2, nft *common.Nft) {
	content.Name = strings.ToLower(content.Name)
	reg := s.ns.GetNameRegisterInfo(content.Name)
	if reg == nil {
		common.Log.Warnf("IndexerMgr.handleNameRouting: %s, Name %s not exist", nft.Base.InscriptionId, content.Name)
		return
	}

	// TODO
	// 只需要当前owner持有该nft就可以修改，而不必在sat上继续铸造
	// if nft.OwnerAddressId != reg.Nft.OwnerAddressId {
	// 	common.Log.Warnf("IndexerMgr.handleNameRouting: %s, Name %s has different owner", nft.Base.InscriptionId, content.Name)
	// 	return
	// }

}

func (s *IndexerMgr) handleOrd(fields map[int][]byte) {

	var nft *common.Nft
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
	case "btcname":
		commonContent := common.ParseCommonContent(string(fields[common.FIELD_CONTENT]))
		if commonContent != nil {
			switch commonContent.Op {
			case "routing":
				s.handleNameRouting(commonContent, nft)
			}
		}
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
