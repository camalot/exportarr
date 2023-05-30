package config

import (
	"github.com/gookit/validate"
	base_config "github.com/onedr0p/exportarr/internal/config"
)

type SabnzbdConfig struct {
	URL              string `validate:"required|url"`
	ApiKey           string `validate:"required"`
	DisableSSLVerify bool
	ApiRootPath      string 
}

func LoadSabnzbdConfig(conf base_config.Config) (*SabnzbdConfig, error) {
	if conf.ApiRootPath == "" {
		conf.ApiRootPath = "/sabnzbd"
	}
	ret := &SabnzbdConfig{
		URL:              conf.URL,
		ApiKey:           conf.ApiKey,
		DisableSSLVerify: conf.DisableSSLVerify,
		ApiRootPath:      conf.ApiRootPath,
	}
	return ret, nil
}

func (c *SabnzbdConfig) Validate() error {
	v := validate.Struct(c)
	if !v.Validate() {
		return v.Errors
	}
	return nil
}
