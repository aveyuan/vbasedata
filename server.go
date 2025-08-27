package vbasedata

type Http struct {
	Addr    string `yaml:"addr" json:"addr"`
	Timeout string `yaml:"timeout" json:"timeout"`
}

type Grpc struct {
	Addr    string `yaml:"addr" json:"addr"`
	Timeout string `yaml:"timeout" json:"timeout"`
}
