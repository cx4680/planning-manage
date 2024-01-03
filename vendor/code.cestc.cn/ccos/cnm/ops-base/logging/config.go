package logging

type LogConfig struct {
	Level              string `yaml:"level"`
	ServiceName        string `yaml:"serviceName"`
	Rotate             string `yaml:"rotate"`
	AccessRotate       string `yaml:"accessRotate"`
	AccessLog          string `yaml:"accessLog"`
	BusinessLog        string `yaml:"businessLog"`
	ServerLog          string `yaml:"serverLog"`
	StatLog            string `yaml:"statLog"`
	ErrorLog           string `yaml:"errLog"`
	LogPath            string `yaml:"logPath"`
	BalanceLogLevel    string `yaml:"balanceLogLevel"`
	GenLogLevel        string `yaml:"genLogLevel"`
	AccessLogOff       bool   `yaml:"accessLogOff"`
	BusinessLogOff     bool   `yaml:"businessLogOff"`
	RequestBodyLogOff  bool   `yaml:"requestLogOff"`
	RespBodyLogMaxSize int    `yaml:"responseLogMaxSize"` // -1:不限制;默认1024字节;
	SuccessStatCode    []int  `yaml:"successStatCode"`
	StorageDay         int64  `yaml:"storageDay"` // 日志保留时间
	WithFile           bool   `yaml:"withFile"`
	MaxSize            int64  `yaml:"maxSize"`
}
