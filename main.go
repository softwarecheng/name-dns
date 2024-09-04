package main

import (
	"github.com/OLProtocol/ordx/common"
	"github.com/OLProtocol/ordx/main/flag"
	"github.com/OLProtocol/ordx/main/g"
)

func init() {
	flag.ParseCmdParams()

	err := g.InitRpc()
	if err != nil {
		common.Log.Fatal(err)
	}
	g.InitSigInt()
}

func main() {
	common.Log.Info("Starting...")
	defer func() {
		g.ReleaseRes()
		common.Log.Info("shut down")
	}()

	err := g.InitBaseIndexer()
	if err != nil {
		common.Log.Error(err)
		return
	}

	err = g.InitRpcService()
	if err != nil {
		common.Log.Error(err)
		return
	}

	// blocked in this thread
	g.RunBaseIndexer()

	common.Log.Info("prepare to release resource...")
}
