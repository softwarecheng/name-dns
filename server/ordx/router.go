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
	r.GET("/ns/status", s.handle.getNSStatus)
	r.GET("/ns/name/:name", s.handle.getNameInfo)
	r.GET("/ns/routing/:name", s.handle.getNameRouting)
}
