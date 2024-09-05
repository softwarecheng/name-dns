package g

import (
	"os"

	"github.com/OLProtocol/ordx/indexer"
)

var (
	SigInt         chan os.Signal
	sigIntFuncList = []func(){}
)

var (
	IndexerMgr *indexer.IndexerMgr
)
