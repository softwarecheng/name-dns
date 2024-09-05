package g

import (
	"fmt"
	"path/filepath"

	common "github.com/OLProtocol/ordx/common"
	"github.com/OLProtocol/ordx/indexer"
	mainCommon "github.com/OLProtocol/ordx/main/common"
	shareBaseIndexer "github.com/OLProtocol/ordx/share/base_indexer"
	"github.com/btcsuite/btcd/chaincfg"
)

func InitBaseIndexer() error {
	periodFlushToDB := int(0)
	if mainCommon.YamlCfg != nil {
		periodFlushToDB = mainCommon.YamlCfg.BasicIndex.PeriodFlushToDB
	} else if mainCommon.Cfg != nil {
		periodFlushToDB = mainCommon.Cfg.PeriodFlushToDB
	} else {
		return fmt.Errorf("cfg is not set")
	}
	chain, err := mainCommon.GetChain()
	if err != nil {
		return err
	}
	chainParam := &chaincfg.MainNetParams
	switch chain {
	case common.ChainTestnet4:
		chainParam = &chaincfg.TestNet3Params
		chainParam.Name = common.ChainTestnet4
	case common.ChainMainnet:
		chainParam = &chaincfg.MainNetParams
	default:
		return fmt.Errorf("unsupported chain: %s", chain)
	}
	dbDir := ""
	if mainCommon.YamlCfg != nil {
		dbDir = mainCommon.YamlCfg.DB.Path
	} else if mainCommon.Cfg != nil {
		dbDir = mainCommon.Cfg.DataDir
	} else {
		return fmt.Errorf("cfg is not set")
	}
	if !filepath.IsAbs(dbDir) {
		dbDir = filepath.Clean(dbDir) + string(filepath.Separator)
	}

	IndexerMgr = indexer.NewIndexerMgr(dbDir, chainParam)
	shareBaseIndexer.InitBaseIndexer(IndexerMgr)
	IndexerMgr.Init()

	if periodFlushToDB != 0 {
		common.Log.WithField("periodFlushToDB", periodFlushToDB).Info("using periodFlushToDB from conf")
		IndexerMgr.WithPeriodFlushToDB(periodFlushToDB)
	}
	return nil
}

func RunBaseIndexer() error {
	stopChan := make(chan bool)
	cb := func() {
		common.Log.Info("handle SIGINT for close base indexer")
		stopChan <- true
	}
	registSigIntFunc(cb)
	common.Log.Info("base indexer start...")
	IndexerMgr.StartDaemon(stopChan)
	return nil
}
