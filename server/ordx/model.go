package ordx

import (
	"github.com/OLProtocol/ordx/share/base_indexer"
)

type Model struct {
	indexer base_indexer.Indexer
}

func NewModel(indexer base_indexer.Indexer) *Model {
	return &Model{
		indexer: indexer,
	}
}
