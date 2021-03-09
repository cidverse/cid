package config

import (
	"github.com/jinzhu/configor"
	"github.com/rs/zerolog/log"
)

func LoadConfigurationFile(config interface{}, file string) (err error) {
	cfgErr := configor.New(&configor.Config{ENVPrefix: "CID", Silent: true}).Load(config, file)

	if cfgErr != nil {
		log.Warn().Str("file",file).Msg("failed to load configuration > " + cfgErr.Error())
	}

	return cfgErr
}