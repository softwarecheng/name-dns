package indexer

import (
	"github.com/OLProtocol/ordx/common"
)

func skipOffsetRange(ord []*common.Range, satpoint int) []*common.Range {
	if satpoint == 0 {
		return ord
	}

	result := make([]*common.Range, 0)
	for _, rng := range ord {
		// skip the offset
		if satpoint > 0 {
			if int64(satpoint) >= (rng.Size) {
				satpoint -= int(rng.Size)
			} else {
				newRange := common.Range{Start: rng.Start + int64(satpoint), Size: rng.Size - int64(satpoint)}
				result = append(result, &newRange)
				satpoint = 0
			}
			continue
		}

		result = append(result, rng)
	}
	return result
}

func reSizeRange(ord []*common.Range, amt int64) []*common.Range {
	result := make([]*common.Range, 0)
	size := int64(0)
	for _, rng := range ord {
		if size+(rng.Size) <= amt {
			result = append(result, rng)
			size += (rng.Size)
		} else {
			newRng := common.Range{Start: rng.Start, Size: (amt - size)}
			result = append(result, &newRng)
			size += (newRng.Size)
		}

		if size == amt {
			break
		}
	}
	return result
}

func reAlignRange(ord []*common.Range, satpoint int, amt int64) []*common.Range {
	ret := skipOffsetRange(ord, satpoint)
	return reSizeRange(ret, amt)
}
