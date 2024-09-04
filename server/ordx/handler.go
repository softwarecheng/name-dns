package ordx

import (
	"net/http"
	"strconv"

	serverOrdx "github.com/OLProtocol/ordx/server/define"
	"github.com/OLProtocol/ordx/share/base_indexer"
	"github.com/gin-gonic/gin"
)

const QueryParamDefaultLimit = "100"

type Handle struct {
	model *Model
}

func NewHandle(indexer base_indexer.Indexer) *Handle {
	return &Handle{
		model: NewModel(indexer),
	}
}

// @Summary Get name service status
// @Description Get name service status
// @Tags ordx
// @Produce json
// @Query start query int false "Start index for pagination"
// @Query limit query int false "Limit for pagination"
// @Security Bearer
// @Success 200 {object} NSStatusResp
// @Failure 401 "Invalid API Key"
// @Router /ns/status [get]
func (s *Handle) getNSStatus(c *gin.Context) {
	resp := &NSStatusResp{
		BaseResp: serverOrdx.BaseResp{
			Code: 0,
			Msg:  "ok",
		},
	}

	start, err := strconv.Atoi(c.DefaultQuery("start", "0"))
	if err != nil {
		start = 0
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", QueryParamDefaultLimit))
	if err != nil {
		limit = 0
	}

	result, err := s.model.GetNSStatusList(start, limit)
	if err != nil {
		resp.Code = -1
		resp.Msg = err.Error()
	} else {
		resp.Data = result
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Get name's properties
// @Description Get name's properties
// @Tags ordx
// @Produce json
// @Security Bearer
// @Success 200 {object} NamePropertiesResp
// @Failure 401 "Invalid API Key"
// @Router /ns/name [get]
func (s *Handle) getNameInfo(c *gin.Context) {
	resp := &NamePropertiesResp{
		BaseResp: serverOrdx.BaseResp{
			Code: 0,
			Msg:  "ok",
		},
	}

	name := c.Param("name")
	result, err := s.model.GetNameInfo(name)
	if err != nil {
		resp.Code = -1
		resp.Msg = err.Error()
	} else {
		resp.Data = result
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Handle) getNameRouting(c *gin.Context) {
	resp := &NameRoutingResp{
		BaseResp: serverOrdx.BaseResp{
			Code: 0,
			Msg:  "ok",
		},
	}

	name := c.Param("name")
	result, err := s.model.GetNameRouting(name)
	if err != nil {
		resp.Code = -1
		resp.Msg = err.Error()
	} else {
		resp.Data = result
	}

	c.JSON(http.StatusOK, resp)
}
