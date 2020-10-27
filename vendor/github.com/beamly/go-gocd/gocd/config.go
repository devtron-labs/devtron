package gocd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"gopkg.in/yaml.v2"
)

// ConfigDirectoryPath is the default location of the `.gocdconf` configuration file
const ConfigDirectoryPath = "~/.gocd.conf"

// Environment variables for configuration.
const (
	EnvVarDefaultProfile = "GOCD_DEFAULT_PROFILE"
	EnvVarServer         = "GOCD_SERVER"
	EnvVarUsername       = "GOCD_USERNAME"
	EnvVarPassword       = "GOCD_PASSWORD"
	EnvVarSkipSsl        = "GOCD_SKIP_SSL_CHECK"
)

// Configuration describes a single connection to a GoCD server
type Configuration struct {
	Server       string
	Username     string `yaml:"username,omitempty"`
	Password     string `yaml:"password,omitempty"`
	SkipSslCheck bool   `yaml:"skip_ssl_check,omitempty" survey:"skip_ssl_check"`
}

// LoadConfigByName loads configurations from yaml at the default file location
func LoadConfigByName(name string, cfg *Configuration) (err error) {

	cfgs, err := LoadConfigFromFile()
	if err == nil {
		newCfg, hasCfg := cfgs[name]
		if !hasCfg {
			return fmt.Errorf("could not find configuration profile '%s'", name)
		}

		*cfg = *newCfg
	} else {
		return err
	}

	if server := os.Getenv(EnvVarServer); server != "" {
		cfg.Server = server
	}

	if username := os.Getenv(EnvVarUsername); username != "" {
		cfg.Username = username
	}

	if password := os.Getenv(EnvVarPassword); password != "" {
		cfg.Password = password
	}

	return nil
}

// LoadConfigFromFile on disk and return it as a Configuration item
func LoadConfigFromFile() (cfgs map[string]*Configuration, err error) {
	var b []byte
	cfgs = make(map[string]*Configuration)

	p, err := ConfigFilePath()
	if err != nil {
		return
	}
	if _, err = os.Stat(p); !os.IsNotExist(err) {
		if b, err = ioutil.ReadFile(p); err != nil {
			return
		}

		if err = yaml.Unmarshal(b, &cfgs); err != nil {
			return
		}
	} else {
		return
	}

	return
}

// ConfigFilePath specifies the default path to a config file
func ConfigFilePath() (configPath string, err error) {
	var usr *user.User

	if configPath = os.Getenv("GOCD_CONFIG_PATH"); configPath != "" {
		return
	}

	// @TODO Make it work for windows. Maybe...
	if usr, err = user.Current(); err != nil {
		return
	}

	configPath = strings.Replace(ConfigDirectoryPath, "~", usr.HomeDir, 1)
	return
}
