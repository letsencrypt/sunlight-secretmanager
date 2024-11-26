package main

import (
	"context"
	"flag"
	"log"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/letsencrypt/sunlight-secretmanager/config"
	secrets "github.com/letsencrypt/sunlight-secretmanager/secrets"
)

func main() {
	fs := flag.NewFlagSet("sunlight", flag.ExitOnError)
	configFlag := fs.String("config", "", "Path to YAML config file")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	nameSeedMap, err := config.LoadConfigFromYaml(*configFlag)
	if err != nil {
		log.Fatalf("failed to read or parse config file: [%v], err: [%v]", configFlag, err)
	}

	log.Printf("seeds: %v", nameSeedMap)

	// Uses Config Profile to initialize AWS SDK configuration.
	// Calls FetchSecretsHelper and passes it configured AWS Secrets Manager client.

	ctx := context.Background()

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithSharedConfigProfile(os.Getenv("AWS_PROFILE")))
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}

	returnedKeys, err := secrets.FetchSecrets(ctx, nameSeedMap, cfg)
	if err != nil {
		log.Printf("failed to load AWS config: [%v], err: [%v]", configFlag, err)
	}

	for key := range returnedKeys {
		log.Printf("successfully loaded key %v", key)
	}
}
