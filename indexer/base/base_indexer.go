package base

import (
	"time"

	"github.com/OLProtocol/ordx/common"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/dgraph-io/badger/v4"
	"github.com/sirupsen/logrus"
)

const SyncStatsKey = "syncStats"

type SyncStats struct {
	ChainTip       int    `json:"chainTip"`
	SyncHeight     int    `json:"syncHeight"`
	SyncBlockHash  string `json:"syncBlockHash"`
	ReorgsDetected []int  `json:"reorgsDetected"`
}

type RpcIndexer struct {
	BaseIndexer
}

func NewRpcIndexer(base *BaseIndexer) *RpcIndexer {
	indexer := &RpcIndexer{
		BaseIndexer: *base.Clone(),
	}

	return indexer
}

type BlockProcCallback func(*common.Block)
type UpdateDBCallback func()

type BaseIndexer struct {
	db         *badger.DB
	stats      *SyncStats
	lastHeight int // 内存数据同步区块
	lastHash   string
	blocksChan chan *common.Block

	// 配置参数
	periodFlushToDB  int
	keepBlockHistory int
	chaincfgParam    *chaincfg.Params

	blockprocCB BlockProcCallback
	updateDBCB  UpdateDBCallback
}

const BLOCK_PREFETCH = 12

func NewBaseIndexer(
	basicDB *badger.DB,
	chaincfgParam *chaincfg.Params,
) *BaseIndexer {
	indexer := &BaseIndexer{
		db:               basicDB,
		stats:            &SyncStats{},
		periodFlushToDB:  500,
		keepBlockHistory: 6,
		blocksChan:       make(chan *common.Block, BLOCK_PREFETCH),
		chaincfgParam:    chaincfgParam,
	}
	return indexer
}

func (b *BaseIndexer) Init(cb1 BlockProcCallback, cb2 UpdateDBCallback) {
	b.blockprocCB = cb1
	b.updateDBCB = cb2

	b.reset()
}

func (b *BaseIndexer) reset() {
	b.loadSyncStatsFromDB()
	b.blocksChan = make(chan *common.Block, BLOCK_PREFETCH)
}

// 只保存UpdateDB需要用的数据
func (b *BaseIndexer) Clone() *BaseIndexer {
	startTime := time.Now()
	newInst := NewBaseIndexer(b.db, b.chaincfgParam)

	newInst.lastHash = b.lastHash
	newInst.lastHeight = b.lastHeight
	newInst.stats = b.stats
	newInst.blockprocCB = b.blockprocCB
	newInst.updateDBCB = b.updateDBCB

	common.Log.Infof("BaseIndexer->clone takes %v", time.Since(startTime))
	return newInst
}

func (b *BaseIndexer) WithPeriodFlushToDB(value int) *BaseIndexer {
	b.periodFlushToDB = value
	return b
}

// only call in compiling data
func (b *BaseIndexer) forceUpdateDB() {
	startTime := time.Now()
	b.UpdateDB()
	common.Log.Infof("BaseIndexer.updateBasicDB: cost: %v", time.Since(startTime))
	b.updateDBCB()
	common.Log.Infof("forceUpdateDB sync to height %d", b.stats.SyncHeight)
}

func (b *BaseIndexer) UpdateDB() {
	common.Log.Infof("BaseIndexer->updateBasicDB %d start...", b.lastHeight)

	wb := b.db.NewWriteBatch()
	defer wb.Cancel()

	b.stats.SyncBlockHash = b.lastHash
	b.stats.SyncHeight = b.lastHeight
	err := common.SetDB([]byte(SyncStatsKey), b.stats, wb)
	if err != nil {
		common.Log.Panicf("BaseIndexer.updateBasicDB-> Error setting in db %v", err)
	}

	err = wb.Flush()
	if err != nil {
		common.Log.Panicf("BaseIndexer.updateBasicDB-> Error satwb flushing writes to db %v", err)
	}
}

func (b *BaseIndexer) forceMajeure() {
	common.Log.Info("Graceful shutdown received, flushing db...")
}

// syncToBlock continues from the sync height to the current height
func (b *BaseIndexer) syncToBlock(height int, stopChan chan struct{}) int {
	if b.lastHeight == height {
		common.Log.Infof("BaseIndexer.SyncToBlock-> already synced to block %d", height)
		return 0
	}

	common.Log.WithFields(logrus.Fields{"BaseIndexer.SyncToBlock-> currentHeight": b.lastHeight, "targetHeight": height}).Info("starting sync")

	// if we don't start from precisely this heigh the UTXO index is worthless
	// we need to start from exactly where we left off
	start := b.lastHeight + 1

	periodProcessedTxs := 0
	startTime := time.Now() // Record the start time

	logProgressPeriod := 1

	stopBlockFetcherChan := make(chan struct{})
	go b.spawnBlockFetcher(start, height, stopBlockFetcherChan)

	for i := start; i <= height; i++ {
		select {
		case <-stopChan:
			b.forceMajeure()
			return -1
		default:
			block := <-b.blocksChan

			if block == nil {
				common.Log.Panicf("BaseIndexer.SyncToBlock-> fetch block failed %d", i)
			}
			//common.Log.Infof("BaseIndexer.SyncToBlock-> get block: cost: %v", time.Since(startTime))

			// make sure that we are at the correct block height
			if block.Height != i {
				common.Log.Panicf("BaseIndexer.SyncToBlock-> expected block height %d, got %d", i, block.Height)
			}

			// detect reorgs
			if i > 0 && block.PrevBlockHash != b.lastHash {
				common.Log.WithField("BaseIndexer.SyncToBlock-> height", i).Warn("reorg detected")
				stopBlockFetcherChan <- struct{}{}
				return block.Height
			}

			// Update the sync stats
			b.stats.ChainTip = height
			b.lastHeight = block.Height
			b.lastHash = block.Hash

			b.blockprocCB(block)

			if block.Height%b.periodFlushToDB == 0 && height-block.Height > b.keepBlockHistory {
				b.forceUpdateDB()
			}

			if i%logProgressPeriod == 0 {
				periodProcessedTxs += len(block.Transactions)
				elapsedTime := time.Since(startTime)
				timePerTx := elapsedTime / time.Duration(periodProcessedTxs)
				readableTime := block.Timestamp.Format("2006-01-02 15:04:05")
				common.Log.Infof("processed block %d (%s) with %d transactions took %v (%v per tx)\n", block.Height, readableTime, periodProcessedTxs, elapsedTime, timePerTx)
				startTime = time.Now()
				periodProcessedTxs = 0
			}
			//common.Log.Info("")
		}
	}

	//b.forceUpdateDB()

	common.Log.Infof("BaseIndexer.SyncToBlock-> already synced to block %d-%d\n", b.lastHeight, b.stats.SyncHeight)
	return 0
}

func (b *BaseIndexer) SyncToChainTip(stopChan chan struct{}) int {
	count, err := getBlockCount()
	if err != nil {
		common.Log.Errorf("failed to get block count %v", err)
		return -1
	}

	bRunInStepMode := false
	if bRunInStepMode {
		if count == uint64(b.lastHeight) {
			return 0
		}
		count = uint64(b.lastHeight) + 1
	}

	return b.syncToBlock(int(count), stopChan)
}

func (b *BaseIndexer) loadSyncStatsFromDB() {
	err := b.db.View(func(txn *badger.Txn) error {
		syncStats := &SyncStats{}
		err := common.GetValueFromDB([]byte(SyncStatsKey), txn, syncStats)
		if err == badger.ErrKeyNotFound {
			common.Log.Info("BaseIndexer.LoadSyncStatsFromDB-> No sync stats found in db")
			syncStats.SyncHeight = -1
		} else if err != nil {
			return err
		}
		common.Log.Infof("stats: %v", syncStats)

		if syncStats.ReorgsDetected == nil {
			syncStats.ReorgsDetected = make([]int, 0)
		}

		b.stats = syncStats
		b.lastHash = b.stats.SyncBlockHash
		b.lastHeight = b.stats.SyncHeight

		return nil
	})

	if err != nil {
		common.Log.Panicf("BaseIndexer.LoadSyncStatsFromDB-> Error loading sync stats from db: %v", err)
	}
}

func (b *BaseIndexer) GetSyncHeight() int {
	return b.stats.SyncHeight
}

func (b *BaseIndexer) GetHeight() int {
	return b.lastHeight
}

func (b *BaseIndexer) SetReorgHeight(height int) {
	b.stats.ReorgsDetected = append(b.stats.ReorgsDetected, height)
}

func (b *BaseIndexer) GetBlockHistory() int {
	return b.keepBlockHistory
}
