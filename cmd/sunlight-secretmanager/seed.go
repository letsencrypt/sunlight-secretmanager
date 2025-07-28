package main

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// All Sunlight seeds must be exactly 32 bytes.
const seedLen = 32

// SecretsManager defines the subset of methods we need from the AWS Secrets Manager
// or equivalent implementation. This makes it easier to mock for testing.
type SecretsManager interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
	CreateSecret(ctx context.Context, params *secretsmanager.CreateSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error)
}

// fetchSeed retrieves a secret value from the provided SecretsManager.
func fetchSeed(ctx context.Context, smClient SecretsManager, id string) ([]byte, error) {
	req := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(id),
		VersionStage: nil,
		VersionId:    nil,
	}

	res, err := smClient.GetSecretValue(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("retrieving secret %q: %w", id, err)
	}

	return res.SecretBinary, nil
}

// createSeed generates a new 32-byte random string and stores it as a new
// secret with the given ID.
func createSeed(ctx context.Context, smClient SecretsManager, id string) ([]byte, error) {
	// crypto/rand.Read is documented to always succeed.
	seed := make([]byte, seedLen)
	_, _ = rand.Read(seed)

	req := &secretsmanager.CreateSecretInput{
		Name:                        aws.String(id),
		AddReplicaRegions:           nil,
		ClientRequestToken:          nil,
		Description:                 nil,
		ForceOverwriteReplicaSecret: false,
		KmsKeyId:                    nil,
		SecretBinary:                seed,
		SecretString:                nil,
		Tags:                        nil,
	}

	_, err := smClient.CreateSecret(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("creating secret %q: %w", id, err)
	}

	return seed, nil
}
