package g

import (
	"github.com/OLProtocol/ordx/common"
	mainCommon "github.com/OLProtocol/ordx/main/common"
	"github.com/OLProtocol/ordx/server"
	serverCommon "github.com/OLProtocol/ordx/server/define"
)

func InitRpcService() error {
	if IndexerMgr == nil {
		return nil
	}
	chain := ""
	maxIndexHeight := int64(0)
	addr := ""

	logPath := ""

	if mainCommon.YamlCfg != nil {
		maxIndexHeight = mainCommon.YamlCfg.BasicIndex.MaxIndexHeight
		rpcService, err := serverCommon.ParseRpcService(mainCommon.YamlCfg.RPCService)
		if err != nil {
			return err
		}
		addr = rpcService.Addr

		logPath = rpcService.LogPath

	} else if mainCommon.Cfg != nil {
		maxIndexHeight = mainCommon.Cfg.MaxIndexHeight
		addr = mainCommon.Cfg.RpcAddr
		logPath = mainCommon.Cfg.RpcLogPath

	}

	chain, err := mainCommon.GetChain()
	if err != nil {
		return err
	}
	Rpc = server.NewRpc(IndexerMgr, chain)
	if maxIndexHeight <= 0 { // default true. set to false when compiling database.
		err := Rpc.Start(addr, logPath)
		if err != nil {
			return err
		}
		common.Log.Info("rpc started")
	}
	return nil
}
