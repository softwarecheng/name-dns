package ft

import (
	"strings"
	"sync"
	"time"

	"github.com/OLProtocol/ordx/common"
	indexer "github.com/OLProtocol/ordx/indexer/common"
	"github.com/OLProtocol/ordx/indexer/nft"
	"github.com/dgraph-io/badger/v4"
)

type TickInfo struct {
	Id             uint64
	Name           string
	MintInfo       *indexer.RangeRBTree            // mint history: 用于查找某个SatRange是否存在该ticker， Value是RBTreeValue_Mint
	InscriptionMap map[string]*common.MintAbbrInfo // key: inscriptionId
	MintAdded      []*common.Mint
	Ticker         *common.Ticker
}

type HolderAction struct {
	UtxoId    uint64
	AddressId uint64
	Index     int
	Tickers   map[string]*common.TickAbbrInfo
	Action    int // -1 删除; 1 增加
}

type HolderInfo struct {
	AddressId uint64
	Index     int
	Tickers   map[string]*common.TickAbbrInfo // key: ticker, 小写
}

type FTIndexer struct {
	db         *badger.DB
	nftIndexer *nft.NftIndexer

	// 所有必要数据都保存在这几个数据结构中，任何查找数据的行为，必须先通过这几个数据结构查找，再去数据库中读其他数据
	// 禁止直接对外暴露这几个结构的数据，防止被不小心修改
	// 禁止直接遍历holderInfo和utxoMap，因为数据量太大（ord有亿级数据）
	mutex      sync.RWMutex                 // 只保护这几个结构
	tickerMap  map[string]*TickInfo         // ticker -> TickerInfo.  name 小写。 数据由mint数据构造。
	holderInfo map[uint64]*HolderInfo       // utxoId -> holder 用于动态更新ticker的holder数据，需要备份到数据库
	utxoMap    map[string]*map[uint64]int64 // ticker -> utxoId -> 资产数量. 动态数据，跟随Holder变更，需要保存在数据库中。

	// 其他辅助信息
	holderActionList []*HolderAction           // 在同一个block中，状态变迁需要按顺序执行，因为一个utxo会很快被消费掉，变成新的utxo
	tickerAdded      map[string]*common.Ticker // key: ticker
}

func NewOrdxIndexer(db *badger.DB) *FTIndexer {
	return &FTIndexer{
		db: db,
	}
}

func (s *FTIndexer) setDBVersion() {
	err := common.SetRawValueToDB([]byte(ORDX_DB_VER_KEY), []byte(ORDX_DB_VERSION), s.db)
	if err != nil {
		common.Log.Panicf("SetRawValueToDB failed %v", err)
	}
}

func (s *FTIndexer) GetDBVersion() string {
	value, err := common.GetRawValueFromDB([]byte(ORDX_DB_VER_KEY), s.db)
	if err != nil {
		common.Log.Errorf("GetRawValueFromDB failed %v", err)
		return ""
	}

	return string(value)
}

// 只保存UpdateDB需要用的数据
func (s *FTIndexer) Clone() *FTIndexer {
	newInst := NewOrdxIndexer(s.db)

	newInst.holderActionList = make([]*HolderAction, len(s.holderActionList))
	copy(newInst.holderActionList, s.holderActionList)

	newInst.tickerAdded = make(map[string]*common.Ticker, 0)
	for key, value := range s.tickerAdded {
		newInst.tickerAdded[key] = value
	}

	newInst.tickerMap = make(map[string]*TickInfo, 0)
	for key, value := range s.tickerMap {
		if len(value.MintAdded) > 0 {
			tick := TickInfo{}
			tick.Name = value.Name
			tick.MintAdded = make([]*common.Mint, len(value.MintAdded))
			copy(tick.MintAdded, value.MintAdded)
			newInst.tickerMap[key] = &tick
		}
	}

	// 保存holderActionList对应的数据
	newInst.holderInfo = make(map[uint64]*HolderInfo, 0)
	newInst.utxoMap = make(map[string]*map[uint64]int64, 0)
	for _, action := range s.holderActionList {
		if action.Action > 0 {
			value, ok := s.holderInfo[action.UtxoId]
			if ok {
				info := HolderInfo{AddressId: value.AddressId, Tickers: value.Tickers}
				newInst.holderInfo[action.UtxoId] = &info
			} //else {
			// 已经被删除，不存在了
			// common.Log.Panicf("can find utxo %s in holderInfo", action.Utxo)
			//}
		}

		for tickerName := range action.Tickers {
			if action.Action > 0 {
				value, ok := s.utxoMap[tickerName]
				if ok {
					amount, ok := (*value)[action.UtxoId]
					if ok {
						newmap, ok := newInst.utxoMap[tickerName]
						if ok {
							(*newmap)[action.UtxoId] = amount
						} else {
							m := make(map[uint64]int64, 0)
							m[action.UtxoId] = amount
							newInst.utxoMap[tickerName] = &m
						}
					} //else {
					// 已经被删除，不存在了
					// common.Log.Panicf("can find utxo %s in utxoMap", action.Utxo)
					//}
				} //else {
				// 已经被删除，不存在了
				// common.Log.Panicf("can find ticker %s in utxoMap", tickerName)
				//}
			}
		}
	}

	return newInst
}

// update之后，删除原来instance中的数据
func (s *FTIndexer) Subtract(another *FTIndexer) {

	s.holderActionList = s.holderActionList[len(another.holderActionList):]

	for key := range another.tickerAdded {
		delete(s.tickerAdded, key)
	}

	for key, value := range another.tickerMap {
		ticker, ok := s.tickerMap[key]
		if ok {
			ticker.MintAdded = ticker.MintAdded[len(value.MintAdded):]
		}
	}

	// 不需要更新 holderInfo 和 utxoMap
}

// 在系统初始化时调用一次，如果有历史数据的话。一般在NewSatIndex之后调用。
func (s *FTIndexer) InitOrdxIndexer(nftIndexer *nft.NftIndexer) {

	s.nftIndexer = nftIndexer
	height := nftIndexer.GetBaseIndexer().GetSyncHeight()

	startTime := time.Now()
	common.Log.Infof("ordx db version: %s", s.GetDBVersion())
	common.Log.Info("InitOrdxIndexerFromDB ...")

	ticks := s.getTickListFromDB()
	if true {
		s.mutex.Lock()

		s.tickerMap = make(map[string]*TickInfo, 0)
		for _, ticker := range ticks {
			s.tickerMap[strings.ToLower(ticker)] = s.initTickInfoFromDB(ticker)
		}

		s.holderInfo = s.loadHolderInfoFromDB()
		s.utxoMap = s.loadUtxoMapFromDB()

		s.holderActionList = make([]*HolderAction, 0)
		s.tickerAdded = make(map[string]*common.Ticker, 0)

		s.mutex.Unlock()
	}

	s.CheckSelf(height)

	elapsed := time.Since(startTime).Milliseconds()
	common.Log.Infof("InitSatIndexFromDB %d ms\n", elapsed)
}

// 自检。如果错误，将停机
func (s *FTIndexer) CheckSelf(height int) bool {
	// 最后才设置dbver
	s.setDBVersion()

	return true
}
