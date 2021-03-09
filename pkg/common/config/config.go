package config

import "github.com/jinzhu/configor"

func LoadConfigurationFile(config interface{}, file string) (err error) {
	return configor.New(&configor.Config{ENVPrefix: "CID", Silent: true}).Load(&config, file)
}