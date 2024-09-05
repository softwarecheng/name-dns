package indexer

// /// rpc interface, run in mul-thread

func (p *IndexerMgr) GetSyncHeight() int {
	return p.rpcService.GetHeight()
}
