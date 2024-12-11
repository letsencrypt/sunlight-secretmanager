package secrets

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/smithy-go/middleware"
)

// Represents error cases.
var (
	errSecretIDNil    = errors.New("SecretId cannot be nil")
	errSecretNotFound = errors.New("secret not found")
)

// Mock implementation of AWSSecretsManagerAPI interface.
type mockSecretsManagerAPI func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)

// GetSecretValue Implements AWS Secret Manager's GetSecretValue for mock.
func (m mockSecretsManagerAPI) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return m(ctx, params, optFns...)
}

// TestFetchSecrets defines test cases for the FetchSecrets function using mock implementations of Secrets Manager SDK.
func TestFetchSecrets(t *testing.T) {
	t.Parallel()
	t.Run("successful secret retrieval", func(t *testing.T) {
		t.Parallel()

		runTestFetchSecrets(
			t,
			map[string]string{"KEY1": "SECRET_1", "KEY2": "SECRET_2"},
			mockSecretsManagerAPI(func(_ context.Context, params *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
				if params.SecretId == nil {
					return nil, errSecretIDNil
				}

				metadata := middleware.Metadata{}
				metadata.Set("mock", "true")

				mockSecretOutput := &secretsmanager.GetSecretValueOutput{
					Name:           nil,
					SecretBinary:   nil,
					SecretString:   nil,
					ARN:            aws.String("arn:aws:secretsmanager:region:account-id:secret:" + *params.SecretId),
					VersionId:      aws.String("version-id"),
					VersionStages:  []string{"AWSCURRENT"},
					CreatedDate:    aws.Time(time.Now()),
					ResultMetadata: metadata,
				}

				switch *params.SecretId {
				case "SECRET_1":
					mockSecretOutput.Name = aws.String("SECRET_1")
					mockSecretOutput.SecretBinary = []byte{226, 151, 186}

					return mockSecretOutput, nil
				case "SECRET_2":
					mockSecretOutput.Name = aws.String("SECRET_2")
					mockSecretOutput.SecretBinary = []byte{104, 101, 108, 108, 111}

					return mockSecretOutput, nil
				default:
					return nil, errSecretNotFound
				}
			}),
			map[string][]byte{
				"SECRET_1": {226, 151, 186},
				"SECRET_2": {104, 101, 108, 108, 111},
			},
			nil,
		)
	})
	t.Run("secret not found", func(t *testing.T) {
		t.Parallel()
		runTestFetchSecrets(
			t,
			map[string]string{"KEY1": "SECRET_1"},
			mockSecretsManagerAPI(func(_ context.Context, _ *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
				return nil, errSecretNotFound
			}),
			nil,
			errSecretNotFound,
		)
	})
}

// RunTestFetchSecrets is a helper function to TestFetchSecrets.
// It runs tests to verify that correct secrets are fetched or appropriate errors are returned.
func runTestFetchSecrets(
	t *testing.T,
	seeds map[string]string,
	client AWSSecretsManagerAPI,
	expect map[string][]byte,
	expectedErr error,
) {
	t.Helper()

	var err error

	ctx := context.Background()

	secret := &Secrets{
		svc: client,
	}

	returnedKeys := make(map[string][]byte)

	for _, seedValue := range seeds {
		input := &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(seedValue),
			VersionStage: aws.String("AWSCURRENT"),
			VersionId:    nil,
		}

		result, getErr := secret.svc.GetSecretValue(ctx, input)
		if getErr != nil {
			err = getErr

			break
		}

		returnedKeys[*result.Name] = result.SecretBinary
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)

		return
	}

	if len(returnedKeys) != len(expect) {
		t.Errorf("expected %v keys, got %v keys", len(expect), len(returnedKeys))

		return
	}

	for expectKey, expectVal := range expect {
		returnedVal, found := returnedKeys[expectKey]
		if !found {
			t.Errorf("expected key %s not found in returned keys", expectKey)

			continue
		}

		if !bytes.Equal(expectVal, returnedVal) {
			t.Errorf("value mismatch for key %s: expected %v, got %v", expectKey, expectVal, returnedVal)
		}
	}
}
