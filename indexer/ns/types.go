package ns

import (
	"github.com/OLProtocol/ordx/common"
	"github.com/OLProtocol/ordx/indexer/ns/pb"
)

const (
	DB_PREFIX_NAME = "r-" // name  NameRegister
	DB_PREFIX_KV   = "k-" // key-value  KeyValueInDB
	DB_PREFIX_BUCK = "bk-"
)

type NameValueInDB = pb.NameValueInDB

// 由nft维持实时状态
type NameRegister struct {
	Nft  *common.Nft
	Name string
}
