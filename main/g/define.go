package g

import (
	"os"

	"github.com/OLProtocol/ordx/indexer"
	"github.com/OLProtocol/ordx/server"
)


var (
	SigInt         chan os.Signal
	sigIntFuncList = []func(){}
)

var (
	Rpc           *server.Rpc
	IndexerMgr    *indexer.IndexerMgr
)
