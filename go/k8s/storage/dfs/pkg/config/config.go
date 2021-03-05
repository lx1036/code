package config

import (
	"encoding/json"
	"io/ioutil"

	"k8s.io/klog/v2"
)

// Config defines the struct of a configuration in general.
type Config struct {
	data map[string]interface{}
	Raw  []byte
}

func (c *Config) parse(fileName string) error {
	jsonFileBytes, err := ioutil.ReadFile(fileName)
	c.Raw = jsonFileBytes
	if err == nil {
		err = json.Unmarshal(jsonFileBytes, &c.data)
	}
	return err
}

// GetString returns a string for the config key.
func (c *Config) GetString(key string) string {
	x, present := c.data[key]
	if !present {
		return ""
	}
	if result, isString := x.(string); isString {
		return result
	}
	return ""
}

// LoadConfigFile loads config information from a JSON file.
func LoadConfigFile(filename string) (*Config, error) {
	result := newConfig()
	err := result.parse(filename)
	if err != nil {
		klog.Errorf("error loading config file %s: %s", filename, err)
	}
	return result, err
}

func newConfig() *Config {
	return &Config{
		data: make(map[string]interface{}),
	}
}
