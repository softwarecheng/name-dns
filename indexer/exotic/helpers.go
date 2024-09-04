package exotic

import (
	"encoding/json"

	"github.com/OLProtocol/ordx/common"
)

type Item struct {
	Output string `json:"output"`
	Start  int64  `json:"start"`
	End    int64  `json:"end"`
	Size   int64  `json:"size"`
	Offset int64  `json:"offset"`
	Rarity string `json:"rarity"`
	Name   string `json:"name"`
}

func ReadRangesFromOrdResponse(fileContent string) []*common.Range {
	// Parse the JSON content into a slice of the struct
	var items []Item
	err := json.Unmarshal([]byte(fileContent), &items)
	if err != nil {
		panic(err)
	}

	res := []*common.Range{}
	for _, item := range items {
		r := &common.Range{
			Start: item.Start,
			Size:  item.Size,
		}
		res = append(res, r)
	}

	return res
}

func IsInBlocks(blocks []int, height int) bool {
	for _, b := range blocks {
		if b == height {
			return true
		}
	}
	return false
}

func IsSatInRange(ranges []*common.Range, sat Sat) bool {
	for _, r := range ranges {
		if int64(sat) >= r.Start && int64(sat) < r.Start+r.Size {
			return true
		}
	}
	return false
}

func IsRodarmorRare(s string) bool {
	if s == Mythic || s == Legendary || s == Epic || s == Rare || s == Uncommon {
		return true
	}

	return false
}

