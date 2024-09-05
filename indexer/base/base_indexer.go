package base

import (
	"sync"
	"time"

	"github.com/OLProtocol/ordx/common"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/dgraph-io/badger/v4"
	"github.com/sirupsen/logrus"
)

type UtxoValue struct {
	Utxo    string
	Address *common.ScriptPubKey
	UtxoId  uint64
	Value   int64
}

type AddressStatus struct {
	AddressId uint64
	Op        int // 0 existed; 1 added
}

type BlockProcCallback func(*common.Block)
type UpdateDBCallback func()

type BaseIndexer struct {
	db      *badger.DB
	stats   *SyncStats // 数据库状态
	reCheck bool

	// 需要clone的数据
	blockVector []*common.BlockValueInDB //
	utxoIndex   *common.UTXOIndex
	delUTXOs    []*UtxoValue // utxo->address,utxoid

	addressIdMap map[string]*AddressStatus

	lastHeight int // 内存数据同步区块
	lastHash   string
	lastSats   int64
	////////////

	blocksChan chan *common.Block

	// 配置参数
	periodFlushToDB  int
	keepBlockHistory int
	chaincfgParam    *chaincfg.Params
	maxIndexHeight   int

	blockprocCB BlockProcCallback
	updateDBCB  UpdateDBCallback
}

const BLOCK_PREFETCH = 12

func NewBaseIndexer(
	basicDB *badger.DB,
	chaincfgParam *chaincfg.Params,
	maxIndexHeight int,
) *BaseIndexer {
	indexer := &BaseIndexer{
		db:               basicDB,
		stats:            &SyncStats{},
		periodFlushToDB:  500,
		keepBlockHistory: 6,
		blocksChan:       make(chan *common.Block, BLOCK_PREFETCH),
		chaincfgParam:    chaincfgParam,
		maxIndexHeight:   maxIndexHeight,
	}

	indexer.addressIdMap = make(map[string]*AddressStatus, 0)

	return indexer
}

func (b *BaseIndexer) Init(cb1 BlockProcCallback, cb2 UpdateDBCallback) {
	dbver := b.GetBaseDBVer()
	common.Log.Infof("base db version: %s", b.GetBaseDBVer())
	if dbver != "" && dbver != BASE_DB_VERSION {
		common.Log.Panicf("DB version inconsistent. DB ver %s, but code base %s", dbver, BASE_DB_VERSION)
	}

	b.blockprocCB = cb1
	b.updateDBCB = cb2

	b.reset()
}

func (b *BaseIndexer) reset() {
	b.loadSyncStatsFromDB()

	b.blocksChan = make(chan *common.Block, BLOCK_PREFETCH)

	b.blockVector = make([]*common.BlockValueInDB, 0)
	b.utxoIndex = common.NewUTXOIndex()
	b.delUTXOs = make([]*UtxoValue, 0)
}

// 只保存UpdateDB需要用的数据
func (b *BaseIndexer) Clone() *BaseIndexer {
	startTime := time.Now()
	newInst := NewBaseIndexer(b.db, b.chaincfgParam, b.maxIndexHeight)

	newInst.utxoIndex = common.NewUTXOIndex()
	for key, value := range b.utxoIndex.Index {
		newInst.utxoIndex.Index[key] = value
	}

	newInst.delUTXOs = make([]*UtxoValue, len(b.delUTXOs))
	copy(newInst.delUTXOs, b.delUTXOs)

	newInst.addressIdMap = make(map[string]*AddressStatus)
	for k, v := range b.addressIdMap {
		newInst.addressIdMap[k] = v
	}
	newInst.blockVector = make([]*common.BlockValueInDB, len(b.blockVector))
	copy(newInst.blockVector, b.blockVector)

	newInst.lastHash = b.lastHash
	newInst.lastHeight = b.lastHeight
	newInst.lastSats = b.lastSats
	newInst.stats = b.stats
	newInst.blockprocCB = b.blockprocCB
	newInst.updateDBCB = b.updateDBCB

	common.Log.Infof("BaseIndexer->clone takes %v", time.Since(startTime))

	return newInst
}

func (b *BaseIndexer) Subtract(another *BaseIndexer) {
	for key := range another.utxoIndex.Index {
		delete(b.utxoIndex.Index, key)
	}

	l := len(another.delUTXOs)
	b.delUTXOs = b.delUTXOs[l:]
}

func needMerge(rngs []*common.Range) bool {
	len1 := len(rngs)
	if len1 < 2 {
		return false
	}

	r1 := rngs[0]
	for i := 1; i < len1; i++ {
		r2 := rngs[i]
		if r1.Start+r1.Size == r2.Start {
			return true
		}
		r1 = r2
	}

	return false
}

func (b *BaseIndexer) Repair() {

	changedUtxoMap := make(map[string]*common.UtxoValueInDB)
	b.db.View(func(txn *badger.Txn) error {
		var err error
		prefix := []byte(common.DB_KEY_UTXO)
		itr := txn.NewIterator(badger.DefaultIteratorOptions)
		defer itr.Close()

		startTime2 := time.Now()
		common.Log.Infof("calculating in %s table ...", common.DB_KEY_UTXO)

		for itr.Seek([]byte(prefix)); itr.ValidForPrefix([]byte(prefix)); itr.Next() {
			item := itr.Item()
			var value common.UtxoValueInDB
			err = item.Value(func(data []byte) error {
				//return common.DecodeBytes(data, &value)
				return common.DecodeBytesWithProto3(data, &value)
			})
			if err != nil {
				common.Log.Panicf("item.Value error: %v", err)
			}

			if needMerge(value.Ordinals) {
				key := item.KeyCopy(nil)
				changedUtxoMap[string(key)] = &value
			}
		}

		common.Log.Infof("%s table %d takes %v", common.DB_KEY_UTXO, len(changedUtxoMap), time.Since(startTime2))
		return nil
	})

	wb := b.db.NewWriteBatch()
	defer wb.Cancel()

	for k, v := range changedUtxoMap {
		v.Ordinals = appendRanges(nil, v.Ordinals)
		err := common.SetDBWithProto3([]byte(k), v, wb)
		if err != nil {
			common.Log.Panicf("Error setting in db %v", err)
		}
	}

	err := wb.Flush()
	if err != nil {
		common.Log.Panicf("BaseIndexer.updateBasicDB-> Error satwb flushing writes to db %v", err)
	}
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

	// startTime = time.Now()
	b.updateDBCB()
	// common.Log.Infof("BaseIndexer.updateOrdxDB: cost: %v", time.Since(startTime))

	common.Log.Infof("forceUpdateDB sync to height %d", b.stats.SyncHeight)
}

func (b *BaseIndexer) closeDB() {
	err := b.db.Close()
	if err != nil {
		common.Log.Errorf("BaseIndexer.closeDB-> Error closing sat db %v", err)
	}
}

func (b *BaseIndexer) prefechAddress() map[string]*common.AddressValueInDB {
	// 测试下提前取的所有地址
	addressValueMap := make(map[string]*common.AddressValueInDB)

	// 在循环次数300万级别时，时间大概1分钟。尽可能不要多次循环这些变量，特别是不要跟updateBasicDB执行通用的操作
	b.db.View(func(txn *badger.Txn) error {
		//startTime := time.Now()

		for _, v := range b.utxoIndex.Index {
			if v.Address.Type == txscript.NullDataTy {
				// 只有OP_RETURN 才不记录
				if v.Value == 0 {
					continue
				}
			}
			b.addUtxo(&addressValueMap, v)
		}

		for _, value := range b.delUTXOs {
			b.removeUtxo(&addressValueMap, value, txn)
		}

		//common.Log.Infof("BaseIndexer.prefechAddress add %d, del %d, address %d in %v\n",
		//	len(b.utxoIndex.Index), len(b.delUTXOs), len(addressValueMap), time.Since(startTime))

		return nil
	})

	return addressValueMap
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

	// reset memory buffer
	b.blockVector = make([]*common.BlockValueInDB, 0)
	b.utxoIndex = common.NewUTXOIndex()
	b.delUTXOs = make([]*UtxoValue, 0)
	b.addressIdMap = make(map[string]*AddressStatus)
}

func (b *BaseIndexer) removeUtxo(addrmap *map[string]*common.AddressValueInDB, utxo *UtxoValue, txn *badger.Txn) {
	utxoId := utxo.UtxoId
	key := common.GetUtxoIdKey(utxoId)
	_, err := txn.Get(key)
	bExist := err == nil
	for _, address := range utxo.Address.Addresses {
		value, ok := (*addrmap)[address]
		if ok {
			if bExist {
				// 存在数据库中，等会去删除
				value.Utxos[utxoId] = &common.UtxoValue{Op: -1}
			} else {
				// 仅从缓存数据中删除
				delete(value.Utxos, utxoId)
			}
		} else {
			if bExist {
				// 存在数据库中，等会去删除
				utxos := make(map[uint64]*common.UtxoValue)
				utxos[utxoId] = &common.UtxoValue{Op: -1}

				id, op := b.getAddressId(address)
				if op >= 0 {
					value = &common.AddressValueInDB{
						AddressType: uint32(utxo.Address.Type),
						AddressId:   id,
						Op:          op,
						Utxos:       utxos,
					}
					(*addrmap)[address] = value
				} else {
					common.Log.Panicf("utxo %x exists but address %s not exists.", utxoId, address)
				}
			}
		}
	}
}

func (b *BaseIndexer) addUtxo(addrmap *map[string]*common.AddressValueInDB, output *common.Output) {
	utxoId := common.GetUtxoId(output)
	sats := output.Value
	for _, address := range output.Address.Addresses {
		value, ok := (*addrmap)[address]
		if ok {
			utxovalue, ok := value.Utxos[utxoId]
			if ok {
				if utxovalue.Value != sats {
					utxovalue.Value = sats
					utxovalue.Op = 1
				}
			} else {
				value.Utxos[utxoId] = &common.UtxoValue{Op: 1, Value: sats}
			}
		} else {
			utxos := make(map[uint64]*common.UtxoValue)
			utxos[utxoId] = &common.UtxoValue{Op: 1, Value: sats}
			id, op := b.getAddressId(address)
			value = &common.AddressValueInDB{
				AddressType: uint32(output.Address.Type),
				AddressId:   id,
				Op:          op,
				Utxos:       utxos,
			}
			(*addrmap)[address] = value
		}
	}
}

func (b *BaseIndexer) forceMajeure() {
	common.Log.Info("Graceful shutdown received, flushing db...")

	b.closeDB()
}

func (b *BaseIndexer) handleReorg(currentBlock *common.Block) {
	common.Log.Warnf("BaseIndexer.handleReorg-> reorg detected at heigh %d", currentBlock.Height)

	// clean memory and reload stats from DB
	// b.reset()
	//b.stats.ReorgsDetected = append(b.stats.ReorgsDetected, currentBlock.Height)
	b.drainBlocksChan()
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
		if b.maxIndexHeight > 0 && b.lastHeight >= b.maxIndexHeight {
			b.forceUpdateDB()
			break
		}

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
				b.handleReorg(block)
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

// 确保输出是第一个。只需要检查第一组的最后一个和第二组的第一个
func appendRanges(rngs1, rngs2 []*common.Range) []*common.Range {
	var r1, r2 *common.Range
	len1 := len(rngs1)
	len2 := len(rngs2)
	if len1 > 0 {
		if len2 == 0 {
			return rngs1
		}
		r1 = rngs1[len1-1]
		r2 = rngs2[0]
		if r1.Start+r1.Size == r2.Start {
			r1.Size += r2.Size
			rngs1 = append(rngs1, rngs2[1:]...)
		} else {
			rngs1 = append(rngs1, rngs2...)
		}
		return rngs1
	} else {
		rngs1 = append(rngs1, rngs2...)
		return rngs1
	}
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
		common.Log.Infof("Code Ver: %s", common.ORDX_INDEXER_VERSION)
		common.Log.Infof("DB Ver: %s", b.GetBaseDBVer())

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

// 耗时很长。仅用于在数据编译完成时验证数据，或者测试时验证数据。
func (b *BaseIndexer) CheckSelf() bool {

	common.Log.Info("BaseIndexer->checkSelf ... ")
	// for height, leak := range b.leakBlocks.SatsLeakBlocks {
	// 	common.Log.Infof("block %d leak %d", height, leak)
	// }
	// common.Log.Infof("Total leaks %d", b.leakBlocks.TotalLeakSats)

	startTime := time.Now()

	lsm, vlog := b.db.Size()
	common.Log.Infof("DB lsm: %0.2f, vlog: %0.2f", float64(lsm)/(1024*1024), float64(vlog)/(1024*1024))

	common.Log.Infof("stats: %v", b.stats)
	common.Log.Infof("Code Ver: %s", common.ORDX_INDEXER_VERSION)
	common.Log.Infof("DB Ver: %s", b.GetBaseDBVer())
	totalSats := common.FirstOrdinalInTheory(b.stats.SyncHeight + 1)
	common.Log.Infof("expected total sats %d", totalSats)

	var wg sync.WaitGroup
	wg.Add(3)

	go b.db.View(func(txn *badger.Txn) error {
		defer wg.Done()

		startTime2 := time.Now()
		common.Log.Infof("calculating in %s table ...", common.DB_KEY_BLOCK)
		var preValue *common.BlockValueInDB
		for i := 0; i <= b.stats.SyncHeight; i++ {
			key := common.GetBlockDBKey(i)
			value := common.BlockValueInDB{}
			err := common.GetValueFromDB(key, txn, &value)
			if err != nil {
				common.Log.Panicf("GetValueFromDB %s error: %v", key, err)
			}
			if value.Height != i {
				common.Log.Panicf("block %d invalid value %d", i, value.Height)
			}
			if preValue != nil {
				if preValue.Ordinals.Start+preValue.Ordinals.Size != value.Ordinals.Start {
					common.Log.Panicf("block %d invalid range %d-%d, %d", i, preValue.Ordinals.Start, preValue.Ordinals.Size, value.Ordinals.Start)
				}
			}

			preValue = &value
		}
		common.Log.Infof("%s table takes %v", common.DB_KEY_BLOCK, time.Since(startTime2))
		return nil
	})

	satsInUtxo := int64(0)
	utxoCount := 0
	nonZeroUtxo := 0
	addressInUtxo := 0
	addressesInT1 := make(map[uint64]bool, 0)
	utxosInT1 := make(map[uint64]bool, 0)
	go b.db.View(func(txn *badger.Txn) error {
		defer wg.Done()

		var err error
		prefix := []byte(common.DB_KEY_UTXO)
		itr := txn.NewIterator(badger.DefaultIteratorOptions)
		defer itr.Close()

		startTime2 := time.Now()
		common.Log.Infof("calculating in %s table ...", common.DB_KEY_UTXO)

		for itr.Seek([]byte(prefix)); itr.ValidForPrefix([]byte(prefix)); itr.Next() {
			item := itr.Item()
			var value common.UtxoValueInDB
			err = item.Value(func(data []byte) error {
				//return common.DecodeBytes(data, &value)
				return common.DecodeBytesWithProto3(data, &value)
			})
			if err != nil {
				common.Log.Panicf("item.Value error: %v", err)
			}

			// 用于打印不存在table2中的utxo
			// if value.UtxoId == 0x17453400960000 {
			// 	key := item.Key()
			// 	str, _ := common.GetUtxoByDBKey(key)
			// 	common.Log.Infof("%x %s", value.UtxoId, str)
			// }

			sats := (common.GetOrdinalsSize(value.Ordinals))
			if sats > 0 {
				nonZeroUtxo++
			}

			satsInUtxo += sats
			utxoCount++

			for _, addressId := range value.AddressIds {
				addressesInT1[addressId] = true
			}
			utxosInT1[value.UtxoId] = true
		}

		addressInUtxo = len(addressesInT1)

		common.Log.Infof("%s table takes %v", common.DB_KEY_UTXO, time.Since(startTime2))
		common.Log.Infof("1. utxo: %d(%d), sats %d, address %d", utxoCount, nonZeroUtxo, satsInUtxo, addressInUtxo)

		return nil
	})

	satsInAddress := int64(0)
	allAddressCount := 0
	allutxoInAddress := 0
	nonZeroUtxoInAddress := 0
	addressesInT2 := make(map[uint64]bool, 0)
	utxosInT2 := make(map[uint64]bool, 0)
	go b.db.View(func(txn *badger.Txn) error {
		defer wg.Done()

		startTime2 := time.Now()
		common.Log.Infof("calculating in %s table ...", common.DB_KEY_ADDRESSVALUE)

		prefix := []byte(common.DB_KEY_ADDRESSVALUE)
		itr := txn.NewIterator(badger.DefaultIteratorOptions)
		defer itr.Close()
		for itr.Seek(prefix); itr.ValidForPrefix(prefix); itr.Next() {
			item := itr.Item()
			value := int64(0)
			item.Value(func(data []byte) error {
				value = int64(common.BytesToUint64(data))
				return nil
			})

			addressId, utxoId, _, index, err := common.ParseAddressIdKey(string(item.Key()))
			if err != nil {
				common.Log.Panicf("ParseAddressIdKey %s failed: %v", string(item.Key()), err)
			}

			allutxoInAddress++
			if index == 0 {
				satsInAddress += value
				if value > 0 {
					nonZeroUtxoInAddress++
				}
			}

			addressesInT2[addressId] = true
			utxosInT2[utxoId] = true
		}
		allAddressCount = len(addressesInT2)

		common.Log.Infof("%s table takes %v", common.DB_KEY_ADDRESSVALUE, time.Since(startTime2))
		common.Log.Infof("2. utxo: %d(%d), sats %d, address %d", allutxoInAddress, nonZeroUtxoInAddress, satsInAddress, allAddressCount)

		return nil
	})

	wg.Wait()

	common.Log.Infof("utxos not in table %s", common.DB_KEY_ADDRESSVALUE)
	utxos1 := findDifferentItems(utxosInT1, utxosInT2)
	if len(utxos1) > 0 {
		ids := b.printfUtxos(utxos1)
		b.deleteUtxos(ids)
		// 因为badger数据库的bug，在DB_KEY_UTXO中删除的数据可能还会出现，在检查后需要重新删除，再次检查，但只重新检查一次
		if !b.reCheck {
			b.reCheck = true
			return b.CheckSelf()
		}
	}

	common.Log.Infof("utxos not in table %s", common.DB_KEY_UTXO)
	utxos2 := findDifferentItems(utxosInT2, utxosInT1)
	if len(utxos2) > 0 {
		ids := b.printfUtxos(utxos2)
		b.deleteUtxos(ids)
	}

	common.Log.Infof("address not in table %s", common.DB_KEY_ADDRESSVALUE)
	utxos3 := findDifferentItems(addressesInT1, addressesInT2)
	for uid := range utxos3 {
		str, _ := common.GetAddressByID(b.db, uid)
		common.Log.Infof("%s", str)
	}

	common.Log.Infof("address not in table %s", common.DB_KEY_UTXO)
	utxos4 := findDifferentItems(addressesInT2, addressesInT1)
	for uid := range utxos4 {
		str, _ := common.GetAddressByID(b.db, uid)
		common.Log.Infof("%s", str)
	}

	if len(utxos1) > 0 || len(utxos2) > 0 || len(utxos3) > 0 || len(utxos4) > 0 {
		common.Log.Panic("utxos or address differents")
	}

	if addressInUtxo != allAddressCount {
		common.Log.Panicf("address count different %d %d", addressInUtxo, allAddressCount)
	}

	if satsInUtxo != satsInAddress {
		common.Log.Panicf("sats different %d %d", satsInAddress, satsInUtxo)
	}

	if nonZeroUtxo != nonZeroUtxoInAddress {
		common.Log.Panicf("utxo different %d %d", nonZeroUtxo, nonZeroUtxoInAddress)
	}

	b.setDBVersion()

	common.Log.Infof("DB checked successfully, %v", time.Since(startTime))
	return true
}

func findDifferentItems(map1, map2 map[uint64]bool) map[uint64]bool {
	differentItems := make(map[uint64]bool)
	for key := range map1 {
		if _, exists := map2[key]; !exists {
			differentItems[key] = true
		}
	}

	return differentItems
}

// only for test
func (b *BaseIndexer) printfUtxos(utxos map[uint64]bool) map[uint64]string {
	result := make(map[uint64]string)
	b.db.View(func(txn *badger.Txn) error {
		var err error
		prefix := []byte(common.DB_KEY_UTXO)
		itr := txn.NewIterator(badger.DefaultIteratorOptions)
		defer itr.Close()

		for itr.Seek([]byte(prefix)); itr.ValidForPrefix([]byte(prefix)); itr.Next() {
			item := itr.Item()
			var value common.UtxoValueInDB
			err = item.Value(func(data []byte) error {
				return common.DecodeBytesWithProto3(data, &value)
			})
			if err != nil {
				common.Log.Errorf("item.Value error: %v", err)
				continue
			}

			// 用于打印不存在table2中的utxo
			if _, ok := utxos[value.UtxoId]; ok {
				key := item.Key()
				str, err := common.GetUtxoByDBKey(key)
				if err == nil {
					common.Log.Infof("%x %s %d", value.UtxoId, str, common.GetOrdinalsSize(value.Ordinals))
					result[value.UtxoId] = str
				}

				delete(utxos, value.UtxoId)
				if len(utxos) == 0 {
					return nil
				}
			}
		}

		return nil
	})

	return result
}

// only for test
func (b *BaseIndexer) deleteUtxos(utxos map[uint64]string) {
	wb := b.db.NewWriteBatch()
	defer wb.Cancel()

	for utxoId, utxo := range utxos {
		key := common.GetUTXODBKey(utxo)
		err := wb.Delete([]byte(key))
		if err != nil {
			common.Log.Errorf("BaseIndexer.updateBasicDB-> Error deleting db: %v\n", err)
		} else {
			common.Log.Infof("utxo deled: %s", utxo)
		}

		err = common.UnBindUtxoId(utxoId, wb)
		if err != nil {
			common.Log.Errorf("BaseIndexer.updateBasicDB-> Error deleting db: %v\n", err)
		} else {
			common.Log.Infof("utxo unbind: %d", utxoId)
		}
	}

	err := wb.Flush()
	if err != nil {
		common.Log.Panicf("BaseIndexer.updateBasicDB-> Error satwb flushing writes to db %v", err)
	}
}

func (b *BaseIndexer) setDBVersion() {
	err := common.SetRawValueToDB([]byte(BaseDBVerKey), []byte(BASE_DB_VERSION), b.db)
	if err != nil {
		common.Log.Panicf("Error setting in db %v", err)
	}
}

func (b *BaseIndexer) GetBaseDBVer() string {
	value, err := common.GetRawValueFromDB([]byte(BaseDBVerKey), b.db)
	if err != nil {
		common.Log.Errorf("GetRawValueFromDB failed %v", err)
		return ""
	}

	return string(value)
}

func (b *BaseIndexer) GetBaseDB() *badger.DB {
	return b.db
}

func (b *BaseIndexer) GetSyncHeight() int {
	return b.stats.SyncHeight
}

func (b *BaseIndexer) GetHeight() int {
	return b.lastHeight
}

func (b *BaseIndexer) GetChainTip() int {
	return b.stats.ChainTip
}

func (b *BaseIndexer) SetReorgHeight(height int) {
	b.stats.ReorgsDetected = append(b.stats.ReorgsDetected, height)
}

func (b *BaseIndexer) GetBlockHistory() int {
	return b.keepBlockHistory
}

func (b *BaseIndexer) ResetBlockVector() {
	b.blockVector = make([]*common.BlockValueInDB, 0)
}

func (p *BaseIndexer) GetBlockInBuffer(height int) *common.BlockValueInDB {
	for _, block := range p.blockVector {
		if block.Height == height {
			return block
		}
	}

	return nil
}

func (p *BaseIndexer) getAddressId(address string) (uint64, int) {
	value, ok := p.addressIdMap[address]
	if !ok {
		common.Log.Errorf("can't find addressId %s", address)
		return common.INVALID_ID, -1
	}
	return value.AddressId, value.Op
}
