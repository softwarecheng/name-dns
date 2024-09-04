package conf

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/OLProtocol/ordx/common"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func LoadConf(cfgPath string) (*Conf, error) {
	confFile, err := os.Open(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cfg: %s, error: %s", cfgPath, err.Error())
	}
	defer confFile.Close()

	conf, err := godotenv.Parse(confFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cfg: %s, error: %s", cfgPath, err.Error())
	}

	bitcoinRPCPort, err := strconv.Atoi(conf["BITCOIN_RPC_PORT"])
	if err != nil {
		common.Log.Fatalln("Error converting BITCOIN_RPC_PORT to int")
	}

	ll := conf["LOG_LEVEL"]
	if ll == "" {
		ll = "info"
	}
	logLevel, err := logrus.ParseLevel(ll)
	if err != nil {
		logLevel = logrus.InfoLevel
	}

	logPath := conf["LOG_PATH"]
	if logPath == "" {
		logPath = "log"
	}

	periodFlushToDB := 500
	periodFlush := conf["PERIOD_FLUSH_TO_DB"]
	if periodFlush != "" {
		periodFlushToDB, err = strconv.Atoi(periodFlush)
		if err != nil {
			common.Log.Fatalln("Error converting PERIOD_FLUSH_TO_DB to int")
		}
	}

	maxIndexHeight, err := strconv.ParseInt(conf["MAX_INDEX_HEIGHT"], 10, 64)
	if err != nil || maxIndexHeight <= 0 {
		maxIndexHeight = -2
	}

	dbDir := conf["DB_DIR"]
	if dbDir == "" {
		dbDir = "./db/"
	}
	dbDir = filepath.FromSlash(dbDir)
	if dbDir[len(dbDir)-1] != filepath.Separator {
		dbDir += string(filepath.Separator)
	}

	rpcAddr := conf["RPC_ADDR"]
	if rpcAddr == "" {
		rpcAddr = "0.0.0.0:8004"
	}

	rpcProxy := conf["RPC_PROXY"]
	if rpcProxy == "" {
		rpcProxy = "/"
	}
	if rpcProxy[0] != '/' {
		rpcProxy = "/" + rpcProxy
	}

	rpcLogPath := conf["RPC_LOG_PATH"]

	swaggerHost := conf["SWAGGER_HOST"]
	if swaggerHost == "" {
		swaggerHost = "127.0.0.1"
	}

	swaggerSchemes := conf["SWAGGER_SCHEMES"]
	if swaggerSchemes == "" {
		swaggerSchemes = "http"
	}

	return &Conf{
		BitCoinChain:    conf["BITCOIN_CHAIN"],
		BitcoinRPCUser:  conf["BITCOIN_RPC_USER"],
		BitcoinRPCPass:  conf["BITCOIN_RPC_PASSWORD"],
		BitcoinRPCPort:  bitcoinRPCPort,
		BitcoinRPCHost:  conf["BITCOIN_RPC_HOST"],
		DataDir:         dbDir,
		LogLevel:        logLevel,
		LogPath:         logPath,
		PeriodFlushToDB: periodFlushToDB,
		MaxIndexHeight:  maxIndexHeight,
		RpcAddr:         rpcAddr,
		RpcLogPath:      rpcLogPath,
	}, nil
}
