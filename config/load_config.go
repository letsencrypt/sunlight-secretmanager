package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"	
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

/*
func LoadAWSConfig(seeds map[string]string) ([]string, error) {

	returnedKeys := []string{}

	config, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("AdministratorAccess-654654394563"))
	if err != nil {
		log.Printf("error with loading default config, %v", err)
		log.Fatal(err)
	}

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(config)

	for seedKey, seedValue := range seeds {

		input := &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(seedValue),
			VersionStage: aws.String("AWSCURRENT"), 
		}

		result, err := svc.GetSecretValue(context.TODO(), input)
		if err != nil {
			log.Printf("input: %v, result: %v, secretName: %v", input, result, *input.SecretId)
			log.Fatal(err.Error())
		}

		// Decrypts secret using the associated KMS key. Commented out because currently not useful
		// var secretString string = *result.SecretString

		log.Printf(*result.Name)
		returnedKeys = append(returnedKeys, seedKey)
	}
	return returnedKeys, nil

}
	*/

type secretsManagerInterface interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// SecretsManagerAPI defines the interface for the AWS Secrets Manager operations required by LoadAWSConfig
type SecretsManagerAPI interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

func LoadAWSConfig(seeds map[string]string) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("AdministratorAccess-654654394563"))
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}

	svc := secretsmanager.NewFromConfig(cfg)
	return LoadSecrets(seeds, svc)
}

func LoadSecrets(seeds map[string]string, api SecretsManagerAPI) ([]string, error) {
	returnedKeys := []string{}

	for seedKey, seedValue := range seeds {
		input := &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(seedValue),
			VersionStage: aws.String("AWSCURRENT"),
		}

		result, err := api.GetSecretValue(context.TODO(), input)
		if err != nil {
			log.Printf("input: %v, secretName: %v", input, *input.SecretId)
			return nil, err
		}

		log.Printf(*result.Name)
		returnedKeys = append(returnedKeys, seedKey)
	}

	return returnedKeys, nil
}