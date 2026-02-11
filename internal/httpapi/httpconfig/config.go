package httpconfig

import (
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server ServerConfig `toml:"server"`
}

type ServerConfig struct {
	Port     string   `toml:"port"`
	ExecPath string   `toml:"exec_path"`
	Timeout  duration `toml:"timeout"`
}

type duration time.Duration

func (d *duration) UnmarshalText(text []byte) error {
	res, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = duration(res)
	return nil
}

func (d duration) Duration() time.Duration {
	return time.Duration(d)
}

func LoadConfig(path string) (*Config, error) {
	var config Config
	// Set defaults
	config.Server.Port = "60204"
	config.Server.ExecPath = `/app/bin/share-sniffer-cli`
	config.Server.Timeout = duration(60 * time.Second)

	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, return default
			return &config, nil
		}
		return nil, err
	}

	// Fallback for docker environment if ExecPath is empty
	if config.Server.ExecPath == "" {
		if _, err := os.Stat("/app/bin/share-sniffer-cli"); err == nil {
			config.Server.ExecPath = "/app/bin/share-sniffer-cli"
		}
	}

	return &config, nil
}
