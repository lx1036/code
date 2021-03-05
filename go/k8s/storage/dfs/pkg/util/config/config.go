package config

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"

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

// GetStringWithDefault returns a default value if key not present
func (c *Config) GetStringWithDefault(key string, defaultVal string) string {
	x, present := c.data[key]
	if !present {
		return defaultVal
	}
	if result, isString := x.(string); isString {
		return result
	}
	return ""
}

// GetInt returns a int value for the config key.
func (c *Config) GetInt(key string) int {
	x, present := c.data[key]
	if !present {
		return 0
	}
	if result, isInt := x.(int); isInt {
		return result
	}
	if result, isInt := x.(int64); isInt {
		return int(result)
	}
	if result, isString := x.(string); isString {
		r, err := strconv.Atoi(result)
		if err == nil {
			return r
		}
	}
	return 0
}

// GetInt64WithDefault returns a int64 value for the config key, if not
// present, return defVal instead
func (c *Config) GetInt64WithDefault(key string, defVal int64) int64 {
	_, present := c.data[key]
	if !present {
		return defVal
	}

	return c.GetInt64(key)
}

// GetBoolWithDefault returns a bool value for the config key.
func (c *Config) GetBoolWithDefault(key string, defaultVal bool) bool {
	x, present := c.data[key]
	if !present {
		return defaultVal
	}
	if result, isBool := x.(bool); isBool {
		return result
	}
	if result, isString := x.(string); isString {
		if result == "true" {
			return true
		}
	}
	return false
}

// GetInt64 returns a int64 value for the config key.
func (c *Config) GetInt64(key string) int64 {
	x, present := c.data[key]
	if !present {
		return 0
	}
	if result, isInt := x.(int64); isInt {
		return result
	}
	if result, isFloat := x.(float64); isFloat {
		return int64(result)
	}
	if result, isString := x.(string); isString {
		r, err := strconv.ParseInt(result, 10, 64)
		if err == nil {
			return r
		}
	}
	return 0
}

// GetDuration returns a int64 value for the config key.
func (c *Config) GetDuration(key string) time.Duration {
	return time.Duration(c.GetInt64(key))
}

// GetBool returns a bool value for the config key.
func (c *Config) GetBool(key string) bool {
	x, present := c.data[key]
	if !present {
		return false
	}
	if result, isBool := x.(bool); isBool {
		return result
	}
	if result, isString := x.(string); isString {
		if result == "true" {
			return true
		}
	}
	return false
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
