package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// LoadFile reads a YAML config file at path, expands ${ENV_VAR} placeholders,
// and returns a Config seeded with DefaultConfig values.
//
// Field mapping uses yaml struct tags so snake_case YAML keys (e.g. buffer_size)
// correctly populate CamelCase Go fields (e.g. BufferSize). Viper's decode hooks
// handle Duration strings ("30s", "5m") automatically.
func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	expanded := os.ExpandEnv(string(data))

	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(expanded)); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := v.Unmarshal(cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "yaml"
	}); err != nil {
		return nil, fmt.Errorf("decoding config: %w", err)
	}
	return cfg, nil
}
