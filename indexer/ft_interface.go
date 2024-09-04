package indexer

func (b *IndexerMgr) GetOrdxDBVer() string {
	return b.ftIndexer.GetDBVersion()
}
