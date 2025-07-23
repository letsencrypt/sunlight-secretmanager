package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// config is a subset of Sunlight's Config. We use it to load config information
// about each log.
type config struct {
	Logs []logConfig
}

// logConfig is a subset of Sunlight's LogConfig. We use it to load just the
// info we need about each individual log.
type logConfig struct {
	// Name is the unique human-readable identifier of the log. We use it for
	// logging purposes.
	Name string
	// Inception is the date at which the log will begin functioning. If the
	// Inception date is in the future, and the Secret retrieved from AWS is
	// empty, then sunlight-secretmanager will create a new secret. If the
	// Inception date is in the past, then sunlight-secretmanager assumes that a
	// previous run should have created a secret, and an empty secret will be
	// treated as an error.
	Inception string
	// Secret is the path to the file where the sunlight instance expects to
	// find this log's secret seed. We write the seed to this path.
	Secret string
}

// loadConfig takes path to a yaml file and returns the seeds in that log file.
// Exported for use in main.go.
func loadConfig(configFile string) (*config, error) {
	yml, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", configFile, err)
	}

	var sunlightConfig config
	err = yaml.Unmarshal(yml, &sunlightConfig)
	if err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", configFile, err)
	}

	if len(sunlightConfig.Logs) == 0 {
		return nil, fmt.Errorf("no logs found in config file %q", configFile)
	}

	for _, log := range sunlightConfig.Logs {
		if log.Name == "" || log.Inception == "" || log.Secret == "" {
			return nil, fmt.Errorf("incomplete config for log %q in config file %q", log.Name, configFile)
		}
	}

	return &sunlightConfig, nil
}
