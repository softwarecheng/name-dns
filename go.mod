module github.com/OLProtocol/ordx

go 1.22.1

// custom versions that add testnet4 support
replace github.com/btcsuite/btcd => github.com/sat20-labs/btcd v0.24.3

replace github.com/btcsuite/btcwallet => github.com/sat20-labs/btcwallet v0.16.11

require (
	github.com/OLProtocol/go-bitcoind v0.0.0-20240716001842-eaea89a7c02d
	github.com/andybalholm/brotli v1.1.0
	github.com/btcsuite/btcd v0.24.2
	github.com/btcsuite/btcd/btcutil v1.1.6
	github.com/dgraph-io/badger/v4 v4.3.0
	github.com/emirpasic/gods v1.18.1
	github.com/fxamacker/cbor/v2 v2.7.0
	github.com/gin-contrib/cors v1.7.2
	github.com/gin-contrib/logger v1.1.2
	github.com/gin-gonic/gin v1.10.0
	github.com/joho/godotenv v1.5.1
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/rs/zerolog v1.33.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.9.0
	github.com/vmihailenco/msgpack/v5 v5.4.1
	google.golang.org/protobuf v1.34.2
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/btcsuite/btcd/btcec/v2 v2.3.4 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	golang.org/x/crypto v0.26.0 // indirect
)

require (
	github.com/aead/siphash v1.0.1 // indirect
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f // indirect
	github.com/bytedance/sonic v1.11.8 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/decred/dcrd/crypto/blake256 v1.0.1 // indirect
	github.com/dgraph-io/ristretto v0.1.2-0.20240116140435-c67e07994f91 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.4 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.22.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/flatbuffers v1.12.1 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kkdai/bstream v1.0.0 // indirect
	github.com/klauspost/compress v1.17.9
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lestrrat-go/strftime v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
