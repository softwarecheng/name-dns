package ns

import (
	"strings"

	"github.com/dgraph-io/badger/v4"
)

func (p *NameService) GetNameRegisterInfo(name string) *NameRegister {
	name = strings.ToLower(name)
	reg := p.getNameInBuffer(name)
	if reg != nil {
		// nft 可能已经被转移了，更新属性

		// reg.Nft
		return reg
	}

	value := NameValueInDB{}
	err := p.db.View(func(txn *badger.Txn) error {
		return loadNameFromDB(name, &value, txn)
	})
	if err != nil {
		return nil
	}

	// nft := p.nftIndexer.GetNftWithId(value.NftId)
	// if nft == nil {
	// 	common.Log.Errorf("GetNftWithId %d failed.", value.NftId)
	// 	return nil
	// }

	reg = &NameRegister{Nft: nil, Name: value.Name}

	return reg
}

// 按照铸造时间
func (p *NameService) GetNames(start, limit int) []string {
	result := make([]string, 0)
	// buckDB := NewBuckStore(p.db)
	// end := start + limit
	// namemap := buckDB.BatchGet(start, end)
	// for _, reg := range p.nameAdded {
	// 	namemap[int(reg.Id)] = &BuckValue{Name: reg.Name, Sat: reg.Nft.Base.Sat}
	// }

	// for i := start; i < end; i++ {
	// 	value, ok := namemap[i]
	// 	if ok {
	// 		result = append(result, value.Name)
	// 	}
	// }

	return result
}
