package common

type KeyValueInDB struct {
	Value         string
	InscriptionId string
}

type NameInfo struct {
	Base *InscribeBaseContent
	Id   int64
	Name string
	KVs  map[string]*KeyValueInDB
}
