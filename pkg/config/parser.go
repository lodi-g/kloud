package config

import (
	"errors"
	"io/ioutil"
	"strings"

	"kloud/pkg/consts"

	"gopkg.in/yaml.v2"
)

type tls struct {
	Enabled bool   `yaml:"enabled"`
	File    string `yaml:"file"`
}

// Config represents the configuration structure
type Config struct {
	Server  string `yaml:"server"`
	ShareID string `yaml:"share"`
}

// Errors thrown by the ValidateConfig func
var (
	ErrMissingScheme = errors.New("missing scheme (http or https) in server")
)

func parseConfig(configFilePath string, config *Config) error {
	in, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(in, &config)
	if err != nil {
		return err
	}

	return nil
}

func validateConfig(config Config) error {
	if strings.HasPrefix(config.Server, "https") == false &&
		strings.HasPrefix(config.Server, "http") == false {
		return ErrMissingScheme
	}

	return nil
}

// Get parses and validate the configuration before retuning it to the caller
func Get() (config Config, err error) {
	if err := parseConfig(consts.InternalDir+"/config.yml", &config); err != nil {
		return Config{}, err
	}

	if err := validateConfig(config); err != nil {
		return Config{}, err
	}

	return config, nil
}
