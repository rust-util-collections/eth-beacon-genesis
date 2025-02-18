package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/ethpandaops/eth-beacon-genesis/config/presets"
)

type Config struct {
	values map[string]string
	preset map[string]string
}

func LoadConfig(path string) (*Config, error) {
	config := &Config{
		values: make(map[string]string),
		preset: make(map[string]string),
	}

	// load config from yaml
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config.values); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	// load referenced preset
	presetName, found := config.GetString("PRESET_BASE")
	if !found || presetName == "" {
		return nil, fmt.Errorf("preset not found")
	}

	presetData, err := presets.PresetsFS.ReadFile(presetName + ".yaml")
	if err != nil {
		return nil, fmt.Errorf("preset '%v' not found: %w", presetName, err)
	}

	if err := yaml.Unmarshal(presetData, &config.preset); err != nil {
		return nil, fmt.Errorf("failed to parse preset yaml: %w", err)
	}

	return config, nil
}

func (c *Config) Get(key string) (string, bool) {
	value, ok := c.values[key]

	if !ok {
		value, ok = c.preset[key]
	}

	return value, ok
}

func (c *Config) GetString(key string) (string, bool) {
	value, ok := c.Get(key)
	if !ok {
		return "", false
	}

	return value, true
}

func (c *Config) GetUint(key string) (uint64, bool) {
	value, ok := c.Get(key)
	if !ok {
		return 0, false
	}

	if val, err := strconv.ParseUint(value, 10, 64); err == nil {
		return val, true
	}

	return 0, false
}

func (c *Config) GetUintDefault(key string, defaultVal uint64) uint64 {
	value, ok := c.GetUint(key)
	if !ok {
		return defaultVal
	}

	return value
}

func (c *Config) GetBytes(key string) ([]byte, bool) {
	value, ok := c.Get(key)
	if !ok {
		return nil, false
	}

	bytes, err := hex.DecodeString(strings.Replace(value, "0x", "", -1))
	if err != nil {
		return nil, false
	}

	return bytes, true
}

func (c *Config) GetBytesDefault(key string, defaultVal []byte) []byte {
	value, ok := c.GetBytes(key)
	if !ok {
		return defaultVal
	}

	return value
}

func (c *Config) GetSpecs() map[string]interface{} {
	specs := make(map[string]interface{})

	for k, v := range c.preset {
		specs[k] = v
	}

	for k, v := range c.values {
		specs[k] = v
	}

	return specs
}
