package define

type RPCService struct {
	Addr    string  `yaml:"addr"`
	Proxy   string  `yaml:"proxy"`
	LogPath string  `yaml:"log_path"`
	Swagger Swagger `yaml:"swagger"`
	API     API     `yaml:"api"`
}

type Swagger struct {
	Host    string   `yaml:"host"`
	Schemes []string `yaml:"schemes"`
}

type API struct {
	APIKeyList     []APIKeyList `yaml:"apikey_list"`
	NoLimitAPIList []string     `yaml:"nolimit_api_list"`
}

type APIKeyList struct {
	APIKey    string     `yaml:"api_key"`
	UserName  string     `yaml:"user_name"`
	RateLimit *RateLimit `yaml:"rate_limit"`
}

type RateLimit struct {
	PerSecond int `yaml:"per_second"`
	PerDay    int `yaml:"per_day"`
}

type APIInfo struct {
	UserName  string     `yaml:"user_name"`
	RateLimit *RateLimit `yaml:"rate_limit"`
}
