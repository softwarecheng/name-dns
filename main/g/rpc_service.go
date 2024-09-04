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
	host := ""
	scheme := ""
	proxy := ""
	logPath := ""
	var apiCfgData any
	if mainCommon.YamlCfg != nil {
		maxIndexHeight = mainCommon.YamlCfg.BasicIndex.MaxIndexHeight
		rpcService, err := serverCommon.ParseRpcService(mainCommon.YamlCfg.RPCService)
		if err != nil {
			return err
		}
		addr = rpcService.Addr
		host = rpcService.Swagger.Host
		for _, v := range rpcService.Swagger.Schemes {
			scheme += v + ","
		}
		proxy = rpcService.Proxy
		logPath = rpcService.LogPath
		if len(rpcService.API.APIKeyList) > 0 || len(rpcService.API.NoLimitAPIList) > 0 {
			apiCfgData = rpcService.API
		}
	} else if mainCommon.Cfg != nil {
		maxIndexHeight = mainCommon.Cfg.MaxIndexHeight
		addr = mainCommon.Cfg.RpcAddr
		host = mainCommon.Cfg.SwaggerHost
		scheme = mainCommon.Cfg.SwaggerSchemes
		proxy = mainCommon.Cfg.RpcProxy
		logPath = mainCommon.Cfg.RpcLogPath
		apiCfgData = nil
	}

	chain, err := mainCommon.GetChain()
	if err != nil {
		return err
	}
	Rpc = server.NewRpc(IndexerMgr, chain)
	if maxIndexHeight <= 0 { // default true. set to false when compiling database.
		err := Rpc.Start(addr, host, scheme,
			proxy, logPath, apiCfgData)
		if err != nil {
			return err
		}
		common.Log.Info("rpc started")
	}
	return nil
}
