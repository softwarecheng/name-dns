package ns

import (
	"sync"
	"time"

	"github.com/OLProtocol/ordx/common"

	"github.com/dgraph-io/badger/v4"
)

// 名字注册到几百万几千万后，这个模块的加载速度查找速度
type NameService struct {
	db *badger.DB

	// 用于快速查找，不释放，尽可能降低内存占用
	mutex sync.RWMutex

	// 缓存

	// 状态变迁
	nameAdded []*NameRegister // 保持顺序

}

func NewNameService(db *badger.DB) *NameService {
	ns := &NameService{
		db: db,
	}
	ns.reset()
	return ns
}

// 只能被调用一次
func (p *NameService) Init() {

}

func (p *NameService) reset() {
	p.nameAdded = make([]*NameRegister, 0)

}

func (p *NameService) Clone() *NameService {
	newInst := NewNameService(p.db)

	newInst.nameAdded = make([]*NameRegister, len(p.nameAdded))
	copy(newInst.nameAdded, p.nameAdded)

	return newInst
}

func (p *NameService) Subtract(another *NameService) {
	p.nameAdded = p.nameAdded[len(another.nameAdded):]
}

// 每个Register都调用
func (p *NameService) NameRegister(reg *NameRegister) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.nameAdded = append(p.nameAdded, reg)
}

// 使用utxoMap，效率高很多
func (p *NameService) UpdateTransfer(block *common.Block) {

}

func (p *NameService) getNameInBuffer(name string) *NameRegister {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	for _, reg := range p.nameAdded {
		if reg.Name == name {
			return reg
		}
	}
	return nil
}

// 跟base数据库同步
func (p *NameService) UpdateDB() {
	//common.Log.Infof("NameService->UpdateDB start...")
	startTime := time.Now()

	// buckDB := NewBuckStore(p.db)
	// buckNames := make(map[int]*BuckValue)

	wb := p.db.NewWriteBatch()
	defer wb.Cancel()

	// index: name
	for _, name := range p.nameAdded {
		key := GetNameKey(name.Name)
		value := NameValueInDB{
			NftId: name.Nft.Base.Id,
			Sat:   name.Nft.Base.Sat,
			Name:  name.Name,
		}
		err := common.SetDBWithProto3([]byte(key), &value, wb)
		//err := common.SetDB([]byte(key), &value, wb)
		if err != nil {
			common.Log.Panicf("NameService->UpdateDB Error setting %s in db %v", key, err)
		}

		// buckNames[int(name.Id)] = &BuckValue{Name: name.Name, Sat: name.Nft.Base.Sat}
	}

	// reset memory buffer
	p.nameAdded = make([]*NameRegister, 0)
	common.Log.Infof("NameService->UpdateDB takes %v", time.Since(startTime))
}
