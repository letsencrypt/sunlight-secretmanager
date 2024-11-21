package secrets_test

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/smithy-go/middleware"
	secrets "github.com/letsencrypt/sunlight-secretmanager/aws"
)

// Represent error cases.
var (
	errSecretIDNil    = errors.New("SecretId cannot be nil")
	errSecretNotFound = errors.New("secret not found")
)

// MockSecretsManagerAPI type mocks output and error responses returned from AWS Secrets Manager SDK.
type mockSecretsManagerAPI func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)

// GetSecretValue Implements AWS Secret Manager's GetSecretValue for mock.
func (m mockSecretsManagerAPI) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return m(ctx, params, optFns...)
}

// newMockSecretManagerAPI returns mock implementation of Secrets Manager SDK.
// Simulates fetching secrets based on provided seeds.
func newMockSecretsManagerAPI(_ context.Context, seeds map[string]string) mockSecretsManagerAPI {
	return func(_ context.Context, params *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
		if params.SecretId == nil {
			return nil, errSecretIDNil
		}

		secretID := *params.SecretId
		if val, ok := seeds[secretID]; ok {
			metadata := middleware.Metadata{}
			metadata.Set("mock", "true")

			return &secretsmanager.GetSecretValueOutput{
				Name:           aws.String(secretID),
				SecretString:   aws.String(val),
				SecretBinary:   nil,
				ARN:            aws.String("arn:aws:secretsmanager:region:account-id:secret:" + secretID),
				VersionId:      aws.String("version-id"),
				VersionStages:  []string{"AWSCURRENT"},
				CreatedDate:    aws.Time(time.Now()),
				ResultMetadata: metadata,
			}, nil
		}

		log.Printf("Failed to retrieve secret %s: %v", secretID, errSecretNotFound)

		return nil, errSecretNotFound
	}
}

// TestFetchSecretsHelper tests the FetchSecretsHelper tests FetchSecretsHelper function by using mock implementations of Secrets Manager SDK.
// Verifies that correct secrets are fetched or appropriate errors are returned.
func TestFetchSecretsHelper(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := []struct {
		name   string
		client mockSecretsManagerAPI
		seeds  map[string]string
		expect []string
		err    error
	}{
		{
			name:   "secret retrieved",
			client: newMockSecretsManagerAPI(ctx, map[string]string{"SECRET_1": "value1", "SECRET_2": "value2"}),
			seeds:  map[string]string{"KEY1": "SECRET_1", "KEY2": "SECRET_2"},
			expect: []string{"SECRET_1", "SECRET_2"},
			err:    nil,
		},
		{
			name:   "secret not found",
			client: newMockSecretsManagerAPI(ctx, map[string]string{}),
			seeds:  map[string]string{"KEY1": "SECRET_1"},
			expect: nil,
			err:    errSecretNotFound,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runTestFetchSecretsHelper(ctx, t, tt.client, tt.seeds, tt.expect, tt.err)
		})
	}
}

// runTestFetchSecretsHelper is a helper function to TestFetchSecretsHelper.
// Compares output of FetchSecretsHelper with expected behavior.
func runTestFetchSecretsHelper(ctx context.Context, t *testing.T, client mockSecretsManagerAPI, seeds map[string]string, expect []string, expectedErr error) {
	t.Helper()

	returnedKeys, err := secrets.FetchSecretsHelper(ctx, seeds, client)

	if expectedErr != nil {
		if err == nil || !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	} else {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}

	if len(expect) != len(returnedKeys) {
		t.Errorf("expected %v keys, got %v keys", len(expect), len(returnedKeys))
	}

	for _, key := range expect {
		found := false

		for _, returnedKey := range returnedKeys {
			if key == returnedKey {
				found = true

				break
			}
		}

		if !found {
			t.Errorf("expected key %s not found in returned keys", key)
		}
	}
}
