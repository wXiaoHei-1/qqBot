package utils

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"sync"
)

const configFilePath = "Your configuration file path"

type Config struct {
	AppID           uint64 `yaml:"appid"`
	Token           string `yaml:"token"`
	DashScopeAPIKey string `yaml:"dashScopeAPIKey"`
	Mysql           string `yaml:"mysql"`
}

var (
	ConfigInfo *Config
	once       sync.Once
)

func NewConfig() *Config {
	once.Do(func() {
		content, err := os.ReadFile(configFilePath)
		if err != nil {
			fmt.Printf("err:%s\n", err)
			panic(err)
		}
		var cfg *Config
		if err = yaml.Unmarshal(content, &cfg); err != nil {
			fmt.Printf("err:%s\n", err)
			panic(err)
		}
		ConfigInfo = cfg
	})
	return ConfigInfo
}
