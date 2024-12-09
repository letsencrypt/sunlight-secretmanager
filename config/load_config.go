package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config struct is from Sunlight github: https://github.com/FiloSottile/sunlight/.
// It contains LogConfigs.
type Config struct {
	Logs []LogConfig
}

// LogConfig struct is from Sunlight github: https://github.com/FiloSottile/sunlight/.
// It contains Seeds.
type LogConfig struct {
	// Name is the fully qualified log name for the checkpoint origin line, as a
	// schema-less URL. It doesn't need to be where the log is actually hosted,
	// but that's advisable.
	Name string

	// Seed is the path to a file containing a secret seed from which the log's
	// private keys are derived. The whole file is used as HKDF input.
	//
	// To generate a new seed, run:
	//
	//   $ head -c 32 /dev/urandom > seed.bin
	//
	Seed string
}

// FileType represents a file with its full path and filename.
type FileType struct {
	Fullpath string
	Filename string
}

// Global map to store file information.
var Files = map[string]FileType{}

// AddFile adds a new file to the global Files map.
func AddFile(name, path string) {
	Files[name] = FileType{
		Fullpath: path,
		Filename: filepath.Base(path),
	}
}

func GetFile(name string) (string, string) {
	file := Files[name]
	fullpath := file.Fullpath
	filename := file.Filename

	return fullpath, filename
}

// LoadConfigFromYaml takes path to a yaml file and returns the seeds in that log file.
// Exported for use in main.go.
func LoadConfigFromYaml(configFile string) (map[string]string, error) {
	yml, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %v: %w", configFile, err)
	}

	var sunlightConfig Config

	if err := yaml.Unmarshal(yml, &sunlightConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file %v: %w", configFile, err)
	}

	logs := sunlightConfig.Logs
	nameSeedMap := make(map[string]string)

	for i := range logs {
		name := logs[i].Name
		seed := logs[i].Seed
		AddFile(name, seed)
		nameSeedMap[logs[i].Name] = filepath.Base(seed)
	}

	return nameSeedMap, nil
}
