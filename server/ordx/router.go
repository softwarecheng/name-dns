package ordx

import (
	"github.com/OLProtocol/ordx/share/base_indexer"
	"github.com/gin-gonic/gin"
)

type Service struct {
	handle *Handle
}

func NewService(indexer base_indexer.Indexer) *Service {
	return &Service{
		handle: NewHandle(indexer),
	}
}

func (s *Service) InitRouter(r *gin.Engine) {
	// 名字服务
	r.GET("/ns/status", s.handle.getNSStatus)
	r.GET("/ns/name/:name", s.handle.getNameInfo)
	r.GET("/ns/values/:name/:prefix", s.handle.getNameValues)
	r.GET("/ns/routing/:name", s.handle.getNameRouting)
	r.GET("/ns/address/:address", s.handle.getNamesWithAddress)
	r.GET("/ns/address/:address/:sub", s.handle.getNamesWithAddress)
	r.GET("/ns/sat/:sat", s.handle.getNamesWithSat)
	r.GET("/ns/inscription/:id", s.handle.getNameWithInscriptionId)
}
