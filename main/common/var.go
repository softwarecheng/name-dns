package define

import (
	"github.com/OLProtocol/ordx/common"
	"github.com/OLProtocol/ordx/main/conf"
)

var (
	Cfg     *conf.Conf
	YamlCfg *conf.YamlConf
)

func GetChain() (string, error) {
	chain := ""
	if YamlCfg != nil {
		chain = YamlCfg.Chain
	} else if Cfg != nil {
		chain = Cfg.BitCoinChain
	}

	switch chain {
	case common.ChainTestnet:
		return common.ChainTestnet, nil
	case common.ChainTestnet4:
		return common.ChainTestnet4, nil
	case common.ChainMainnet:
		return common.ChainMainnet, nil
	default:
		return common.ChainTestnet4, nil
	}
}
