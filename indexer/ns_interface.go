package indexer

import (
	"github.com/OLProtocol/ordx/common"
)

func (b *IndexerMgr) GetNameInfo(name string) *common.NameInfo {
	reg := b.ns.GetNameRegisterInfo(name)
	if reg == nil {
		common.Log.Errorf("GetNameRegisterInfo %s failed", name)
		return nil
	}

	return nil
}

func (b *IndexerMgr) GetNames(start, limit int) []string {
	return b.ns.GetNames(start, limit)
}
