package config

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"gopkg.in/yaml.v3"
)

// Struct is from Sunlight github: https://github.com/FiloSottile/sunlight/.
// It contains LogConfigs.
type Config struct {
	Logs []LogConfig
}

// Struct is from Sunlight github: https://github.com/FiloSottile/sunlight/.
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
		nameSeedMap[logs[i].Name] = logs[i].Seed
	}

	return nameSeedMap, nil
}

// SecretsManagerAPI defines the interface for the AWS Secrets Manager operations required by FetchSecrets.
type SecretsManagerAPI interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// Retrieves secrets from AWS Secrets Manager given a name-to-seed mapping.
// Returns list of successfully loadeded keys or error.
func FetchSecrets(seeds map[string]string, cfg aws.Config) ([]string, error) {
	returnedKeys := []string{}

	api := secretsmanager.NewFromConfig(cfg)

	for _, seedValue := range seeds {
		input := &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(seedValue),
			VersionStage: aws.String("AWSCURRENT"),
			VersionId:    nil,
		}

		result, err := api.GetSecretValue(context.TODO(), input)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve secret for %v: %w", *input.SecretId, err)
		}

		returnedKeys = append(returnedKeys, *result.Name)
	}

	return returnedKeys, nil
}
