package secrets

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// AWSSecretsManagerAPI defines the interface for the AWS Secrets Manager operations required by FetchSecretsHelper.
type AWSSecretsManagerAPI interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// FetchSecrets uses Config Profile to initialize AWS SDK configuration.
// Calls FetchSecretsHelper and passes it configured AWS Secrets Manager client.
func FetchSecrets(ctx context.Context, seeds map[string]string, cfg aws.Config) ([]string, error) {
	svc := secretsmanager.NewFromConfig(cfg)

	return FetchSecretsHelper(ctx, seeds, svc)
}

// FetchSecretsHelper retrieves secrets from AWS Secrets Manager given a name-to-seed mapping.
// Returns list of successfully loadeded keys or error.
func FetchSecretsHelper(ctx context.Context, seeds map[string]string, api AWSSecretsManagerAPI) ([]string, error) {
	returnedKeys := []string{}

	for _, seedValue := range seeds {
		input := &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(seedValue),
			VersionStage: aws.String("AWSCURRENT"),
			VersionId:    nil,
		}

		result, err := api.GetSecretValue(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve secret for %v: %w", *input.SecretId, err)
		}

		returnedKeys = append(returnedKeys, *result.Name)
	}

	return returnedKeys, nil
}
