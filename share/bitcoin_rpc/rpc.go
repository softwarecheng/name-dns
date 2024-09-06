package bitcoin_rpc

import (
	"github.com/OLProtocol/go-bitcoind"
)

var ShareBitconRpc *bitcoind.Bitcoind

func InitBitconRpc(host string, port int, user, passwd string, useSSL bool) error {
	var err error
	ShareBitconRpc, err = bitcoind.New(
		host,
		port,
		user,
		passwd,
		useSSL,
		3600, // server timeout is 1 hour for debug
	)
	return err
}
