package base

const SyncStatsKey = "syncStats"
const BaseDBVerKey = "dbver"

// 1.1.0  2024.07.01-
// 1.2.0  2024.07.20    multi-address
const BASE_DB_VERSION = "1.2.0"

type SyncStats struct {
	ChainTip       int    `json:"chainTip"`
	SyncHeight     int    `json:"syncHeight"`
	SyncBlockHash  string `json:"syncBlockHash"`
	ReorgsDetected []int  `json:"reorgsDetected"`
}

type IrregularSubsidy struct {
	TotalLeakSats  int64
	SatsLeakBlocks map[int]int64
}
