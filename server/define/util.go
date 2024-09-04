package define

import (
	"gopkg.in/yaml.v2"
)

func ParseRpcService(data any) (*RPCService, error) {
	rpcServiceRaw, err := yaml.Marshal(data)
	if err != nil {
		return nil, err
	}
	ret := &RPCService{}
	err = yaml.Unmarshal(rpcServiceRaw, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
