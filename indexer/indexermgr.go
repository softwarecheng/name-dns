package indexer

import (
	"sync"
	"time"

	"github.com/OLProtocol/ordx/common"
	base_indexer "github.com/OLProtocol/ordx/indexer/base"

	"github.com/OLProtocol/ordx/indexer/ft"
	"github.com/OLProtocol/ordx/indexer/nft"
	"github.com/OLProtocol/ordx/indexer/ns"

	"github.com/btcsuite/btcd/chaincfg"

	"github.com/dgraph-io/badger/v4"
)

type IndexerMgr struct {
	dbDir string
	// data from blockchain
	baseDB *badger.DB
	ftDB   *badger.DB
	nsDB   *badger.DB
	nftDB  *badger.DB
	// data from market

	// 配置参数
	chaincfgParam   *chaincfg.Params
	ordxFirstHeight int
	ordFirstHeight  int
	maxIndexHeight  int

	ftIndexer *ft.FTIndexer
	ns        *ns.NameService
	nft       *nft.NftIndexer
	clmap     map[common.TickerName]map[string]int64 // collections map, ticker -> inscriptionId -> asset amount

	mutex sync.RWMutex
	// 跑数据
	lastCheckHeight int
	compiling       *base_indexer.BaseIndexer
	// 备份所有需要写入数据库的数据
	compilingBackupDB *base_indexer.BaseIndexer

	ordxBackupDB *ft.FTIndexer
	nsBackupDB   *ns.NameService
	nftBackupDB  *nft.NftIndexer
	// 接收前端api访问的实例，隔离内存访问
	rpcService *base_indexer.RpcIndexer

	// 本地缓存，在区块更新时清空
	addressToNftMap  map[string][]*common.Nft
	addressToNameMap map[string][]*common.Nft
}

var instance *IndexerMgr

func NewIndexerMgr(
	dbDir string,
	chaincfgParam *chaincfg.Params,
	maxIndexHeight int,
) *IndexerMgr {

	if instance != nil {
		return instance
	}

	mgr := &IndexerMgr{
		dbDir:             dbDir,
		chaincfgParam:     chaincfgParam,
		maxIndexHeight:    maxIndexHeight,
		compilingBackupDB: nil,

		ordxBackupDB: nil,
		nsBackupDB:   nil,
		nftBackupDB:  nil,
		rpcService:   nil,
	}

	instance = mgr
	switch instance.chaincfgParam.Name {
	case "mainnet":
		instance.ordFirstHeight = 767430
		instance.ordxFirstHeight = 827307
	case "testnet3":
		instance.ordFirstHeight = 2413343
		instance.ordxFirstHeight = 2570589
	default: // testnet4
		instance.ordFirstHeight = 0
		instance.ordxFirstHeight = 0
	}

	return instance
}

func (b *IndexerMgr) Init() {
	err := b.initDB()
	if err != nil {
		common.Log.Panicf("initDB failed. %v", err)
	}
	b.compiling = base_indexer.NewBaseIndexer(b.baseDB, b.chaincfgParam, b.maxIndexHeight)
	b.compiling.Init(b.processOrdProtocol, b.forceUpdateDB)
	b.lastCheckHeight = b.compiling.GetSyncHeight()

	dbver := b.GetBaseDBVer()
	common.Log.Infof("base db version: %s", dbver)
	if dbver != "" && dbver != base_indexer.BASE_DB_VERSION {
		common.Log.Panicf("DB version inconsistent. DB ver %s, but code base %s", dbver, base_indexer.BASE_DB_VERSION)
	}

	b.nft = nft.NewNftIndexer(b.nftDB)
	b.nft.Init(b.compiling)
	b.ftIndexer = ft.NewOrdxIndexer(b.ftDB)
	b.ftIndexer.InitOrdxIndexer(b.nft)
	b.ns = ns.NewNameService(b.nsDB)
	b.ns.Init(b.nft)

	b.rpcService = base_indexer.NewRpcIndexer(b.compiling)

	b.compilingBackupDB = nil

	b.ordxBackupDB = nil
	b.nsBackupDB = nil
	b.nftBackupDB = nil

	b.addressToNftMap = nil
	b.addressToNameMap = nil
}

func (b *IndexerMgr) GetBaseDB() *badger.DB {
	return b.baseDB
}

func (b *IndexerMgr) WithPeriodFlushToDB(value int) *IndexerMgr {
	b.compiling.WithPeriodFlushToDB(value)
	return b
}

func (b *IndexerMgr) StartDaemon(stopChan chan bool) {
	n := 10
	ticker := time.NewTicker(time.Duration(n) * time.Second)

	stopIndexerChan := make(chan struct{}, 1) // 非阻塞

	if b.repair() {
		common.Log.Infof("repaired, check again.")
		return
	}

	bWantExit := false
	isRunning := false
	disableSync := false
	tick := func() {
		if disableSync {
			return
		}
		if !isRunning {
			isRunning = true
			go func() {
				ret := b.compiling.SyncToChainTip(stopIndexerChan)
				if ret == 0 {
					if b.maxIndexHeight > 0 {
						if b.maxIndexHeight <= b.compiling.GetHeight() {
							b.checkSelf()
							common.Log.Infof("reach expected height, set exit flag")
							bWantExit = true
						}
					} else {
						b.updateDB()
					}
				} else if ret > 0 {
					// handle reorg
					b.handleReorg(ret)
				} else {
					common.Log.Infof("IndexerMgr inner thread exit by SIGINT signal")
					bWantExit = true
				}

				isRunning = false
			}()
		}
	}

	tick()
	for !bWantExit {
		select {
		case <-ticker.C:
			if bWantExit {
				break
			}
			tick()
		case <-stopChan:
			common.Log.Info("IndexerMgr got SIGINT")
			if bWantExit {
				break
			}
			if isRunning {
				select {
				case stopIndexerChan <- struct{}{}:
					// 成功发送
				default:
					// 通道已满或没有接收者，执行其他操作
				}
				for isRunning {
					time.Sleep(time.Second / 10)
				}
				common.Log.Info("IndexerMgr inner thread exited")
			}
			bWantExit = true
		}
	}

	ticker.Stop()

	// close all
	b.closeDB()

	common.Log.Info("IndexerMgr exited.")
}

func (b *IndexerMgr) closeDB() {

	common.RunBadgerGC(b.baseDB)
	common.RunBadgerGC(b.nftDB)
	common.RunBadgerGC(b.nsDB)
	common.RunBadgerGC(b.ftDB)

	b.ftDB.Close()
	b.nsDB.Close()
	b.nftDB.Close()
	b.baseDB.Close()

}

func (b *IndexerMgr) checkSelf() {
	start := time.Now()
	b.compiling.CheckSelf()

	b.nft.CheckSelf(b.baseDB)
	b.ftIndexer.CheckSelf(b.compiling.GetSyncHeight())
	b.ns.CheckSelf(b.baseDB)
	common.Log.Infof("IndexerMgr.checkSelf takes %v", time.Since(start))
}

func (b *IndexerMgr) forceUpdateDB() {
	startTime := time.Now()

	b.nft.UpdateDB()
	b.ns.UpdateDB()
	b.ftIndexer.UpdateDB()

	common.Log.Infof("IndexerMgr.forceUpdateDB: takes: %v", time.Since(startTime))
}

func (b *IndexerMgr) handleReorg(height int) {
	b.closeDB()
	b.Init()
	b.compiling.SetReorgHeight(height)
	common.Log.Infof("IndexerMgr handleReorg completed.")
}

// 为了回滚数据，我们采用这样的策略：
// 假设当前最新高度是h，那么数据库记录，最多只到（h-6），这样确保即使回滚，只需要从数据库回滚即可
// 为了保证数据库记录最高到（h-6），我们做一次数据备份，到合适实际再写入数据库
func (b *IndexerMgr) updateDB() {
	b.updateServiceInstance()

	if b.compiling.GetHeight()-b.compiling.GetSyncHeight() < b.compiling.GetBlockHistory() {
		common.Log.Infof("updateDB do nothing at height %d-%d", b.compiling.GetHeight(), b.compiling.GetSyncHeight())
		return
	}

	if b.compiling.GetHeight()-b.compiling.GetSyncHeight() == b.compiling.GetBlockHistory() {
		// 先备份数据在缓存
		if b.compilingBackupDB == nil {
			b.prepareDBBuffer()
			common.Log.Infof("updateDB clone data at height %d-%d", b.compiling.GetHeight(), b.compiling.GetSyncHeight())
		}
		return
	}

	// 这个区间不备份数据
	if b.compiling.GetHeight()-b.compiling.GetSyncHeight() < 2*b.compiling.GetBlockHistory() {
		common.Log.Infof("updateDB do nothing at height %d-%d", b.compiling.GetHeight(), b.compiling.GetSyncHeight())
		return
	}

	// b.GetHeight()-b.GetSyncHeight() == 2*b.GetBlockHistory()

	// 到达双倍高度时，将备份的数据写入数据库中。
	if b.compilingBackupDB != nil {
		if b.compiling.GetHeight()-b.compilingBackupDB.GetHeight() < b.compiling.GetBlockHistory() {
			common.Log.Infof("updateDB do nothing at height %d, backup instance %d", b.compiling.GetHeight(), b.compilingBackupDB.GetHeight())
			return
		}
		common.Log.Infof("updateDB do backup->forceUpdateDB() at height %d-%d", b.compiling.GetHeight(), b.compiling.GetSyncHeight())
		b.performUpdateDBInBuffer()
	}
	b.prepareDBBuffer()
	common.Log.Infof("updateDB clone data at height %d-%d", b.compiling.GetHeight(), b.compiling.GetSyncHeight())
}

func (b *IndexerMgr) performUpdateDBInBuffer() {
	b.cleanDBBuffer() // must before UpdateDB
	b.compilingBackupDB.UpdateDB()

	b.nftBackupDB.UpdateDB()
	b.nsBackupDB.UpdateDB()
	b.ordxBackupDB.UpdateDB()
}

func (b *IndexerMgr) prepareDBBuffer() {
	b.compilingBackupDB = b.compiling.Clone()
	b.compiling.ResetBlockVector()

	b.ordxBackupDB = b.ftIndexer.Clone()
	b.nsBackupDB = b.ns.Clone()
	b.nftBackupDB = b.nft.Clone()
	common.Log.Infof("backup instance %d cloned", b.compilingBackupDB.GetHeight())
}

func (b *IndexerMgr) cleanDBBuffer() {
	b.compiling.Subtract(b.compilingBackupDB)

	b.ftIndexer.Subtract(b.ordxBackupDB)
	b.ns.Subtract(b.nsBackupDB)
	b.nft.Subtract(b.nftBackupDB)
}

func (b *IndexerMgr) updateServiceInstance() {
	if b.rpcService.GetHeight() == b.compiling.GetHeight() {
		return
	}

	newService := base_indexer.NewRpcIndexer(b.compiling)
	common.Log.Infof("service instance %d cloned", newService.GetHeight())

	newService.UpdateServiceInstance()
	b.mutex.Lock()
	b.rpcService = newService
	b.addressToNftMap = nil
	b.addressToNameMap = nil
	b.mutex.Unlock()
}

func (p *IndexerMgr) repair() bool {
	//p.compiling.Repair()
	return false
}
