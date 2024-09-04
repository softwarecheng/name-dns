package exotic

import "github.com/OLProtocol/ordx/common"

const MAX_EPOCH = 34

// casey's ordinals, sat20 reset all value in running
var startingSats = [MAX_EPOCH]Sat{
	Sat(0),
	Sat(1050000000000000),
	Sat(1575000000000000),
	Sat(1837500000000000),
	Sat(1968750000000000),
	Sat(2034375000000000),
	Sat(2067187500000000),
	Sat(2083593750000000),
	Sat(2091796875000000),
	Sat(2095898437500000),
	Sat(2097949218750000),
	Sat(2098974609270000),
	Sat(2099487304530000),
	Sat(2099743652160000),
	Sat(2099871825870000),
	Sat(2099935912620000),
	Sat(2099967955890000),
	Sat(2099983977420000),
	Sat(2099991988080000),
	Sat(2099995993410000),
	Sat(2099997995970000),
	Sat(2099998997250000),
	Sat(2099999497890000),
	Sat(2099999748210000),
	Sat(2099999873370000),
	Sat(2099999935950000),
	Sat(2099999967240000),
	Sat(2099999982780000),
	Sat(2099999990550000),
	Sat(2099999994330000),
	Sat(2099999996220000),
	Sat(2099999997060000),
	Sat(2099999997480000),
	Sat(common.MaxSupply),
}

type Epoch int64

func SetEpochStartingSat(e int64, s int64) {
	startingSats[e] = Sat(s)
}

func SetEpochStartingAndChangeLast(e int64, s int64) {
	if int64(startingSats[e]) != s {
		underpay := int64(startingSats[e]) - s
		startingSats[e] = Sat(s)
		e++
		for e < MAX_EPOCH {
			startingSats[e] = Sat(int64(startingSats[e]) - underpay)
			e++
		}
	}
}

func (e Epoch) GetStartingSat() Sat {
	return startingSats[e]
}

func EpochFromSat(sat Sat) Epoch {
	for i, e := range startingSats {
		if sat < e {
			return Epoch(i - 1)
		}
	}

	return Epoch(0)
}

func (e Epoch) GetSubsidy() int64 {
	return 50 * 100000000 >> e
}
