package common

import (
	"bytes"
	"encoding/gob"

	badger "github.com/dgraph-io/badger/v4"
)

func SetDB(key []byte, data interface{}, wb *badger.WriteBatch) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return err
	}
	return wb.Set([]byte(key), []byte(buf.Bytes()))
}

func GetValueFromDB(key []byte, txn *badger.Txn, target interface{}) error {
	item, err := txn.Get([]byte(key))
	if err != nil {
		return err
	}
	err = item.Value(func(v []byte) error {
		err := DecodeBytes(v, target)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func DecodeBytes(data []byte, target interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(target)
}

// 不能与其他DB读写混用，要确保这一点
func RunBadgerGC(db *badger.DB) {
	// 只有在跑数据后，打开，压缩数据，同时做严格的数据检查
	if db.IsClosed() {
		return
	}

	for {
		err := db.RunValueLogGC(0.5)
		Log.Infof("RunValueLogGC return %v", err)
		if err == badger.ErrNoRewrite {
			break
		} else if err != nil {
			break
		}
	}
	db.Sync()
	Log.Info("badgerGc: RunValueLogGC is done")

}
