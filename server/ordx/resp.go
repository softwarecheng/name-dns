package ordx

import (
	ordx "github.com/OLProtocol/ordx/common"
	serverOrdx "github.com/OLProtocol/ordx/server/define"
)

type RangesReq struct {
	Ranges []*ordx.Range `json:"ranges"`
}

type NSStatusData struct {
	Version string                `json:"version"`
	Total   uint64                `json:"total"`
	Start   uint64                `json:"start"`
	Names   []*serverOrdx.NftItem `json:"names"`
}

type NSStatusResp struct {
	serverOrdx.BaseResp
	Data *NSStatusData `json:"data"`
}

type NamePropertiesResp struct {
	serverOrdx.BaseResp
	Data *serverOrdx.OrdinalsName `json:"data"`
}

type NameRoutingResp struct {
	serverOrdx.BaseResp
	Data *serverOrdx.NameRouting `json:"data"`
}

type NamesWithAddressData struct {
	Address string                     `json:"address"`
	Total   int                        `json:"total"`
	Names   []*serverOrdx.OrdinalsName `json:"names"`
}

type NamesWithAddressResp struct {
	serverOrdx.BaseResp
	Data *NamesWithAddressData `json:"data"`
}
