package configs

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

type Config interface {
	*ServerConfig | *AgentConfig
}

func processConfigFile[T Config](cfg T) ([]string, error) {
	configFileFlags := flag.NewFlagSet("config", flag.ContinueOnError)
	var c, config string
	configFileFlags.StringVar(&c, "c", "", "path to config file")
	configFileFlags.StringVar(&config, "config", "", "path to config file")
	if err := configFileFlags.Parse(os.Args[1:]); err != nil {
		return os.Args[1:], nil
	}
	envConfigFile, envExists := os.LookupEnv("CONFIG")
	paths := []string{c, config}
	if envExists {
		paths = append(paths, envConfigFile)
	}
	for _, path := range paths {
		if path != "" {
			if err := parseConfigFile(cfg, path); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}
	return configFileFlags.Args(), nil
}

func parseConfigFile[T Config](cfg T, path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		if errClose := f.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()
	data, err := io.ReadAll(f)
	if err != nil {
		return
	}
	return json.Unmarshal(data, cfg)
}
