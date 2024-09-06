package common

import (
	badger "github.com/dgraph-io/badger/v4"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func SetDBWithProto3(key []byte, data protoreflect.ProtoMessage, wb *badger.WriteBatch) error {
	dataBytes, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	return wb.Set([]byte(key), dataBytes)
}

func GetValueFromDBWithProto3(key []byte, txn *badger.Txn, target protoreflect.ProtoMessage) error {
	item, err := txn.Get([]byte(key))
	if err != nil {
		return err
	}
	return item.Value(func(v []byte) error {
		return proto.Unmarshal(v, target)
	})
}
