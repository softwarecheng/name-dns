package bitcoind

import (
	"github.com/gin-gonic/gin"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) InitRouter(r *gin.Engine, basePath string) {
	//broadcast raw tx => blockstream api: POST /tx
	r.POST(basePath+"/btc/tx", s.sendRawTx)
	r.GET(basePath+"/btc/tx/:txid", s.getTx)
	r.GET(basePath+"/btc/rawtx/:txid", s.getRawTx)
	r.GET(basePath+"/btc/block/:blockhash", s.getRawBlock)
	r.GET(basePath+"/btc/block/blockhash/:height", s.getBlockHash)
	r.GET(basePath+"/btc/block/bestblockheight", s.getBestBlockHeight)
}
