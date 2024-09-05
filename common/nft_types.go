package common

import (
	"github.com/OLProtocol/ordx/common/pb"
)

type InscribeBaseContent = pb.InscribeBaseContent
type Nft struct {
	Base           *InscribeBaseContent
	OwnerAddressId uint64
	UtxoId         uint64
}

type NftsInSat = pb.NftsInSat

const ALL_TICKERS = "*"

type TickerName struct {
	TypeName string
	Name     string // * 所有ticker
}
