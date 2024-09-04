package nft

import (
	"github.com/OLProtocol/ordx/common"
	"github.com/dgraph-io/badger/v4"
)

func (p *NftIndexer) HasNftInUtxo(utxoId uint64) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	_, ok := p.utxoMap[utxoId]
	if ok {
		return true
	}

	bHasNft := false
	p.db.View(func(txn *badger.Txn) error {
		bHasNft = hasNftInUtxo(utxoId, txn)
		return nil
	})
	return bHasNft
}

func (p *NftIndexer) GetNftWithInscriptionId(inscriptionId string) *common.Nft {
	if inscriptionId == "" {
		return nil
	}

	nft := p.getNftInBuffer2(inscriptionId)
	if nft != nil {
		return nft
	}

	var value InscriptionInDB
	err := p.db.View(func(txn *badger.Txn) error {
		key := GetInscriptionIdKey(inscriptionId)
		return common.GetValueFromDB([]byte(key), txn, &value)
	})

	if err != nil {
		//common.Log.Errorf("GetValueFromDB with inscription %s failed. %v", inscriptionId, err)
		//return nil
	} else {
		nfts := p.GetNftsWithSat(value.Sat)
		if nfts != nil {
			for _, nft := range nfts.Nfts {
				if nft.Id == value.Id {
					return &common.Nft{
						Base:           nft,
						OwnerAddressId: nfts.OwnerAddressId, UtxoId: nfts.UtxoId}
				}
			}
		}
	}

	return nil
}

func (p *NftIndexer) GetNftHolderWithInscriptionId(inscriptionId string) uint64 {
	nft := p.GetNftWithInscriptionId(inscriptionId)
	if nft != nil {
		return nft.OwnerAddressId
	}
	return common.INVALID_ID
}

// key: sat
func (p *NftIndexer) GetBoundSatsWithUtxo(utxoId uint64) []int64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	value := &NftsInUtxo{}
	p.db.View(func(txn *badger.Txn) error {
		return loadUtxoValueFromDB(utxoId, value, txn)
	})
	//if err != nil {
	// 还没有保存到数据库
	// return nil
	//}

	satmap := make(map[int64]bool)
	for _, sat := range value.Sats {
		satmap[sat] = true
	}

	sats, ok := p.utxoMap[utxoId]
	if ok {
		for sat := range sats {
			satmap[sat] = true
		}
	}

	result := make([]int64, 0)
	for sat := range satmap {
		result = append(result, sat)
	}

	return result
}


func (p *NftIndexer) GetNftsWithUtxo(utxoId uint64) []*common.Nft {
	sats := p.GetBoundSatsWithUtxo(utxoId)

	result := make([]*common.Nft, 0)
	for _, sat := range sats {
		info := p.GetNftsWithSat(sat)
		if info != nil {
			for _, nft := range info.Nfts {
				result = append(result, &common.Nft{Base:nft, 
					OwnerAddressId: info.OwnerAddressId, UtxoId: utxoId})
			}
		}
	}

	return result
}

func (p *NftIndexer) GetNftWithId(id int64) *common.Nft {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	nft := p.getNftInBuffer(id)
	if nft != nil {
		return nft
	}

	buckDB := NewBuckStore(p.db)
	bv, err := buckDB.Get(id)
	if err != nil {
		return nil
	}

	nfts := &common.NftsInSat{}
	err = p.db.View(func(txn *badger.Txn) error {
		return loadNftFromDB(bv.Sat, nfts, txn)
	})
	if err != nil {
		return nil
	}

	for _, nft := range nfts.Nfts {
		if nft.Id == id {
			return &common.Nft{
				Base:           nft,
				OwnerAddressId: nfts.OwnerAddressId, UtxoId: nfts.UtxoId}
		}
	}

	return nil
}

// return sats
func (p *NftIndexer) GetNftsWithRanges(rngs []*common.Range) []int64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	result := make([]int64, 0)
	

	for _, rng := range rngs {
		startKey := []byte(GetSatKey(rng.Start))
		endKey := []byte(GetSatKey(rng.Start + rng.Size - 1))
		err := common.IterateRangeInDB(p.db, startKey, endKey, func(key, value []byte) error {
			sat, err := ParseSatKey(string(key))
			if err == nil {
				result = append(result, sat)
			}
			return err
		})
		if err != nil {
			common.Log.Errorf("IterateRangeInDB failed. %v", err)
		}
	}

	return result
}

func (p *NftIndexer) GetNftsWithSat(sat int64) *common.NftsInSat {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	nfts := &common.NftsInSat{}
	err := p.db.View(func(txn *badger.Txn) error {
		return loadNftFromDB(sat, nfts, txn)
	})

	nft := p.getNftInBuffer4(sat)
	if nft != nil {
		if err != nil {
			nfts.OwnerAddressId = nft.OwnerAddressId
			nfts.Sat = nft.Base.Sat
			nfts.UtxoId = nft.UtxoId
		}
		nfts.Nfts = append(nfts.Nfts, nft.Base)
	}

	return nfts
}

func (p *NftIndexer) GetStatus() *common.NftStatus {
	return p.status
}

// 按照铸造时间
func (p *NftIndexer) GetNfts(start, limit int) ([]int64, int) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	end := start + limit

	result := make([]int64, 0)
	buckDB := NewBuckStore(p.db)
	idmap := buckDB.BatchGet(int64(start), int64(end))
	for _, nft := range p.nftAdded {
		idmap[nft.Base.Id] = &BuckValue{nft.Base.Sat}
	}
	for i := start; i < end; i++ {
		_, ok := idmap[int64(i)]
		if ok {
			result = append(result, int64(i))
		}
	}

	return result, len(idmap)
}

// 按照铸造时间
func (p *NftIndexer) GetNftsWithInscriptionAddress(addressId uint64, start, limit int) ([]int64, int) {
	result := p.GetAllNftsWithInscriptionAddress(addressId)

	total := len(result)
	end := total
	if start >= end {
		return nil, 0
	}
	if start+limit < end {
		end = start + limit
	}

	return result[start:end], total
}

// 按照铸造时间
func (p *NftIndexer) GetAllNftsWithInscriptionAddress(addressId uint64) []int64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	result := getNftsWithAddressFromDB(addressId, p.db)
	for _, nft := range p.nftAdded {
		if nft.Base.InscriptionAddress == addressId {
			result = append(result, nft.Base.Id)
		}
	}

	return result
}
