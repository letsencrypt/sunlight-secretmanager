package main

import (
	"context"
	"flag"
	"log"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/letsencrypt/sunlight-secretmanager/config"
)

func main() {
	fs := flag.NewFlagSet("sunlight", flag.ExitOnError)
	configFlag := fs.String("config", "", "Path to YAML config file")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Println("Error parsing flags", err)
		os.Exit(1)
	}

	c, err := config.LoadConfigFromYaml(*configFlag)
	if err != nil {
		log.Printf("failed to read or parse config file: [%v], err: [%v]", configFlag, err)
	} else {
		log.Printf("seeds: %v", c)
	}

	// Uses Config Profile to initialize AWS SDK configuration.
	// Calls FetchSecrets and passes it configured AWS Secrets Manager client.

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithSharedConfigProfile(os.Getenv("AWS_PROFILE")))

	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}

	returnedKeys, err := config.FetchSecrets(c, cfg)

	if err != nil {
		log.Printf("failed to load AWS config: [%v], err: [%v]", configFlag, err)
	} else {
		for key := range returnedKeys {
			log.Printf("successfully loaded key %v with value %v", key, returnedKeys[key])
		}
	}
}
