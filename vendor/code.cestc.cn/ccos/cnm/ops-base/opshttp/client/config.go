package client

type HttpClientConfig struct {
	ServiceName        string `yaml:"serviceName"`
	Timeout            int64  `yaml:"timeout"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	MaxConn            int    `yaml:"maxConn"`
	MaxIdleConn        int    `yaml:"maxIdleConn"`
	IdleConnTimeout    int    `yaml:"idleConnTimeout"`
}
