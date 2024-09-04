package indexer

import (
	"fmt"
	"strings"

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
	p.baseDB, err = openDB(p.dbDir+"base", opts)
	if err != nil {
		return err
	}
	
	p.nftDB, err = openDB(p.dbDir+"nft", opts)
	if err != nil {
		return err
	}
	
	p.nsDB, err = openDB(p.dbDir+"ns", opts)
	if err != nil {
		return err
	}
	
	p.ftDB, err = openDB(p.dbDir+"ft", opts)
	if err != nil {
		return err
	}
	
	p.localDB, err = openDB(p.dbDir+"local", opts)
	if err != nil {
		return err
	}

	return nil
}

func getCollectionKey(ntype, ticker string) []byte {
	return []byte("c-"+ntype+"-"+ticker)
}

func parseCollectionKey(key string) (string, string, error) {
	parts := strings.Split(key, "-")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("invalid key %s", key)
	}
	return parts[1], parts[2], nil
}

func inscriptionIdsToCollectionMap(ids []string) map[string]int64 {
	inscmap := make(map[string]int64)
	for _, id := range ids {
		inscmap[id] = 1
	}
	return inscmap
}

func (p *IndexerMgr) initCollections() {
	common.Log.Info("initCollections ...")

	p.clmap = make(map[common.TickerName]map[string]int64)
	err := p.localDB.View(func(txn *badger.Txn) error {
		prefixBytes := []byte("c-")
		prefixOptions := badger.DefaultIteratorOptions
		prefixOptions.Prefix = prefixBytes
		it := txn.NewIterator(prefixOptions)
		defer it.Close()
		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			item := it.Item()
			key := string(item.Key())

			nty, name, err := parseCollectionKey(key)
			if err == nil {
				var ids []string
				value, err := item.ValueCopy(nil)
				if err != nil {
					common.Log.Errorln("initCollections ValueCopy " + key + " " + err.Error())
				} else {
					err = common.DecodeBytes(value, &ids)
					if err == nil {
						p.clmap[common.TickerName{TypeName: nty, Name:name}] = inscriptionIdsToCollectionMap(ids)
					} else {
						common.Log.Errorln("initCollections DecodeBytes " + err.Error())
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		common.Log.Panicf("initCollections Error: %v", err)
	}
}
