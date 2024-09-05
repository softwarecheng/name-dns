package base_indexer

import (
	"github.com/OLProtocol/ordx/common"
)

type Indexer interface {

	// NameService
	GetNameInfo(name string) *common.NameInfo
	GetNames(start, limit int) []string
}
