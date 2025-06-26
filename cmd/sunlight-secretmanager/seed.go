package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// SecretsManager defines the subset of methods we need from the AWS Secrets Manager
// or equivalent implementation. This makes it easier to mock for testing.
type SecretsManager interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// fetchSeed retrieves a secret value from the provided SecretsManager.
func fetchSeed(ctx context.Context, smClient SecretsManager, id string) ([]byte, error) {
	req := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(id),
		VersionStage: aws.String("AWSCURRENT"),
		VersionId:    nil,
	}

	res, err := smClient.GetSecretValue(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("retrieving secret %q: %w", id, err)
	}

	return res.SecretBinary, nil
}
