package config

import (
	"errors"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	FileName = ".kube-watcher.yaml"
)

type Webhook struct {
	Url string `json:"url"`
}
type Mail struct {
	From string `json:"from"`
	To string `json:"to"`
}
type Sink struct {
	Webhook Webhook `json:"webhook"`
}
type Config struct {
	Sink Sink `json:"sink" yaml:"sink"`
}

func New() (*Config, error) {
	c := &Config{}
	if err := c.Load(); err != nil {
		return nil, err
	}
	
	c.DefaultVars()

	return c, nil
}

func (config *Config) Load() error {
	file, err := os.Open(getConfigFile())
	if err != nil {
		return err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	if len(content) != 0 {
		return yaml.Unmarshal(content, config)
	}

	return errors.New("config file content is empty")
}

func (config *Config) DefaultVars()  {
	
}

func getConfigFile() string {
	return filepath.Join(configDir(), FileName)
}

func configDir() string {
	configDir := viper.GetString("CONFIG_DIR")
	if len(configDir) != 0 {
		return configDir
	}
	
	return os.Getenv("HOME")
}
