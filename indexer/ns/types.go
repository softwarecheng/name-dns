package ns

import (
	"github.com/OLProtocol/ordx/common"
	"github.com/OLProtocol/ordx/indexer/ns/pb"
)

const NS_DB_VERSION = "1.0.1"
const NS_DB_VERSION_KEY = "nsdbver"
const NS_STATUS_KEY = "nsstatus"

const (
	DB_PREFIX_NAME = "r-" // name  NameRegister
	DB_PREFIX_KV   = "k-" // key-value  KeyValueInDB
	DB_PREFIX_BUCK = "bk-"
)

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TransferAction struct {
	UtxoId    uint64
	AddressId uint64
	Names     map[string]bool
	Action    int // -1 删除; 1 增加
}

type NameValueInDB = pb.NameValueInDB

// 由nft维持实时状态
type NameRegister struct {
	Nft  *common.Nft
	Name string
}

type NameProperties struct {
	NameRegister
}

// 一个聪只能绑定一个名字。再次绑定会报错。
type RBTreeValue_Name struct {
	Name string
}
