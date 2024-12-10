package secrets

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/letsencrypt/sunlight-secretmanager/config"
)

type AWSSecretsManagerAPI interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type Secrets struct {
	svc AWSSecretsManagerAPI
}

func New(cfg aws.Config) *Secrets {
	return &Secrets{
		svc: secretsmanager.NewFromConfig(cfg),
	}
}

// FetchSecrets uses Config Profile to initialize AWS SDK configuration.
// Calls FetchSecretsHelper and passes it configured AWS Secrets Manager client.
func FetchSecrets(ctx context.Context, seeds map[string]string, _ map[string]config.FileType, cfg aws.Config) (map[string][]byte, error) {
	api := New(cfg)

	returnedKeys := make(map[string][]byte)

	for _, seedValue := range seeds {
		input := &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(seedValue),
			VersionStage: aws.String("AWSCURRENT"),
			VersionId:    nil,
		}

		result, err := api.svc.GetSecretValue(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve secret for %v: %w", *input.SecretId, err)
		}

		res := *result
		secretName := res.Name
		secretValue := res.SecretBinary

		returnedKeys[*secretName] = secretValue
	}

	return returnedKeys, nil
}
