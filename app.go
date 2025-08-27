package vbasedata

type App struct {
	AppName string `yaml:"app_name" json:"app_name"`
	Env     string `yaml:"env" json:"env"`
	Health  string `yaml:"health" json:"health"`
	Trace   string `yaml:"trace" json:"trace"`
}
