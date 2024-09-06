package common

const (
	PROTOCOL_NAME = "ordx"
)
const (
	FIELD_CONTENT          = 0
	FIELD_CONTENT_TYPE     = 1
	FIELD_POINT            = 2
	FIELD_PARENT           = 3
	FIELD_META_DATA        = 5
	FIELD_META_PROTOCOL    = 7
	FIELD_CONTENT_ENCODING = 9
	FIELD_DELEGATE         = 11
	FIELD_INVALID1         = 20
	FIELD_INVALID2         = 21
)

const MAX_NAME_LEN = 32
const MIN_NAME_LEN = 3

type OrdxBaseContent struct {
	P  string `json:"p,omitempty"`
	Op string `json:"op,omitempty"`
}

type Brc20BaseContent struct {
	OrdxBaseContent
	Ticker string `json:"tick,omitempty"`
}

type PrimaryNameBaseContent struct {
	OrdxBaseContent
	Name   string `json:"name"`
	Avatar string `json:"avatar,omitempty"`
}

type OrdxRegContent struct {
	OrdxBaseContent
	Name string   `json:"name"`
	KVs  []string `json:"kvs,omitempty"`
}

type OrdxUpdateContentV1 struct {
	OrdxBaseContent
	Name string   `json:"name"`
	KVs  []string `json:"kvs"` // ["key1=value1", "key2=value2", ...]
}

type OrdxUpdateContentV2 struct {
	P    string
	Op   string
	Name string
	KVs  map[string]string
}
