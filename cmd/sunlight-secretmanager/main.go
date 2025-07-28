// sunlight-secretmanager ensures that Sunlight CT Log instances have access to
// their secret key material. In particular, it reads the log's config file,
// extracts the paths at which each log in that config expects to find its
// key material, fetches the corresponding secrets from the AWS Secrets Manager
// API, and writes the secrets to the expected location.
//
// Usage:
//
//	sunlight-secretmanager -config /path/to/config.yaml
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// tmpfsMagic is the magic number used to indicate that a unix filesystem is
// a tmpfs. The value is copied from golang.org/x/sys/unix.TMPFS_MAGIC, which we
// can't use here because its source file has a "go:build linux" directive.
// https://cs.opensource.google/go/x/sys/+/refs/tags/v0.34.0:unix/zerrors_linux.go;l=3498
const tmpfsMagic = 0x1021994

func main() {
	flagset := flag.NewFlagSet("sunlight-secretmanager", flag.ContinueOnError)
	configFlag := flagset.String("config", "", "Path to YAML config file")
	fileSystemFlag := flagset.Int64("filesystem", tmpfsMagic, "OS Filesystem constant to enforce writing to. Defaults to Linux tmpfs")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Error parsing flags: %s", err)
	}

	config, err := loadConfig(*configFlag)
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	ctx := context.Background()

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithSharedConfigProfile(os.Getenv("AWS_PROFILE")))
	if err != nil {
		log.Fatalf("Error loading default AWS config: %s", err)
	}

	smClient := secretsmanager.NewFromConfig(cfg)

	for _, logConf := range config.Logs {
		seed, err := getOrCreateSeed(ctx, filepath.Base(logConf.Secret), logConf.Inception, smClient)
		if err != nil {
			log.Fatalf("Error getting seed for log %q: %v", logConf.Name, err)
		}

		err = writeFile(logConf.Secret, seed, *fileSystemFlag)
		if err != nil {
			log.Fatalf("Error persisting seed for log %q: %v", logConf.Name, err)
		}
	}
}

func getOrCreateSeed(ctx context.Context, id string, inception string, smClient SecretsManager) ([]byte, error) {
	seed, err := fetchSeed(ctx, smClient, id)
	if err != nil {
		return nil, fmt.Errorf("error fetching seed: %w", err)
	}

	if len(seed) == 0 {
		if time.Now().Format(time.DateOnly) != inception {
			return nil, errors.New("log has empty seed, but today is not the Inception date")
		}

		seed, err = createSeed(ctx, smClient, id)
		if err != nil {
			return nil, fmt.Errorf("error creating seed: %w", err)
		}
	}

	return seed, nil
}
