package indexer

import (
	"github.com/OLProtocol/ordx/common"
	"github.com/dgraph-io/badger/v4"
)

func openDB(filepath string, opts badger.Options) (db *badger.DB, err error) {
	opts = opts.WithDir(filepath).WithValueDir(filepath).WithLoggingLevel(badger.WARNING)
	db, err = badger.Open(opts)
	if err != nil {
		return nil, err
	}
	common.Log.Infof("InitDB-> start db gc for %s", filepath)
	common.RunBadgerGC(db)
	return db, nil
}

func (p *IndexerMgr) initDB() (err error) {
	common.Log.Info("InitDB-> start...")
	opts := badger.DefaultOptions("").WithBlockCacheSize(3000 << 20)
	p.nsDB, err = openDB(p.dbDir+"ns", opts)
	if err != nil {
		return err
	}
	return nil
}
