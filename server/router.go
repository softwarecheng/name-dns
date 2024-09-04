package server

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/OLProtocol/ordx/indexer"
	"github.com/OLProtocol/ordx/server/base"
	"github.com/OLProtocol/ordx/server/bitcoind"
	serverCommon "github.com/OLProtocol/ordx/server/define"

	"github.com/OLProtocol/ordx/server/ordx"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	STRICT_TRANSPORT_SECURITY   = "strict-transport-security"
	CONTENT_SECURITY_POLICY     = "content-security-policy"
	CACHE_CONTROL               = "cache-control"
	VARY                        = "vary"
	ACCESS_CONTROL_ALLOW_ORIGIN = "access-control-allow-origin"
	TRANSFER_ENCODING           = "transfer-encoding"
	CONTENT_ENCODING            = "content-encoding"
)

const (
	CONTEXT_TYPE_TEXT = "text/html; charset=utf-8"
	CONTENT_TYPE_JSON = "application/json"
)

type RateLimit struct {
	limit    *limiter.Limiter
	reqCount int
}

type Rpc struct {
	basicService *base.Service
	ordxService  *ordx.Service
	btcdService  *bitcoind.Service

	apiConf      *serverCommon.API
	initApiConf  bool
	apiConfMutex sync.Mutex
	apiLimitMap  sync.Map
}

func NewRpc(baseIndexer *indexer.IndexerMgr, chain string) *Rpc {
	return &Rpc{
		basicService: base.NewService(baseIndexer),
		ordxService:  ordx.NewService(baseIndexer),
		btcdService:  bitcoind.NewService(),
	}
}

func (s *Rpc) Start(rpcUrl, swaggerHost, swaggerSchemes, rpcProxy, rpcLogFile string, apiConf any) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	var writers []io.Writer
	if rpcLogFile != "" {
		exePath, _ := os.Executable()
		executableName := filepath.Base(exePath)
		if strings.Contains(executableName, "debug") {
			executableName = "debug"
		}
		executableName += ".rpc"
		fileHook, err := rotatelogs.New(
			rpcLogFile+"/"+executableName+".%Y%m%d.log",
			rotatelogs.WithLinkName(rpcLogFile+"/"+executableName+".log"),
			rotatelogs.WithMaxAge(24*time.Hour),
			rotatelogs.WithRotationTime(1*time.Hour),
		)
		if err != nil {
			return fmt.Errorf("failed to create RotateFile hook, error %s", err)
		}
		writers = append(writers, fileHook)
	}
	writers = append(writers, os.Stdout)
	gin.DefaultWriter = io.MultiWriter(writers...)
	r.Use(logger.SetLogger(
		logger.WithLogger(logger.Fn(func(c *gin.Context, l zerolog.Logger) zerolog.Logger {
			if c.Request.Header["Authorization"] == nil {
				return l
			}
			return l.With().
				Str("Authorization", c.Request.Header["Authorization"][0]).
				Logger()
		})),
	))

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.OptionsResponseStatusCode = 200
	r.Use(cors.New(config))

	// doc
	s.InitApiDoc(swaggerHost, swaggerSchemes, rpcProxy)
	r.GET(rpcProxy+"/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// api config
	err := s.InitApiConf(apiConf)
	if err != nil {
		return err
	}

	err = s.applyApiConf(r, rpcProxy)
	if err != nil {
		return err
	}

	// common header
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set(VARY, "Origin")
		c.Writer.Header().Add(VARY, "Access-Control-Request-Method")
		c.Writer.Header().Add(VARY, "Access-Control-Request-Headers")

		c.Writer.Header().Del(CONTENT_SECURITY_POLICY)
		c.Writer.Header().Set(
			CONTENT_SECURITY_POLICY,
			"default-src 'self'",
		)

		c.Writer.Header().Set(
			STRICT_TRANSPORT_SECURITY,
			"max-age=31536000; includeSubDomains; preload",
		)

		c.Writer.Header().Set(
			ACCESS_CONTROL_ALLOW_ORIGIN,
			"*",
		)

		c.Next()
	})

	// router
	s.basicService.InitRouter(r, rpcProxy)
	s.ordxService.InitRouter(r, rpcProxy)
	s.btcdService.InitRouter(r, rpcProxy)
	parts := strings.Split(rpcUrl, ":")
	if len(parts) < 2 {
		rpcUrl += ":80"
	}

	go r.Run(rpcUrl)
	return nil
}
