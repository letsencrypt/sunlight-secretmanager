package config_test

import (
	"context"
	"errors"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/smithy-go/middleware"
	"github.com/letsencrypt/sunlight-secretmanager/config"
)

func TestLoadConfigNoFile(t *testing.T) {
	t.Parallel()

	testFile := ""
	got, err := config.LoadConfigFromYaml(testFile)

	if got != nil || err == nil {
		t.Errorf("got %q and error %q, wanted error and nil error", got, err)
	}
}

func TestLoadConfigCorrect(t *testing.T) {
	t.Parallel()

	testFile := "sunlight.yaml"
	got, err := config.LoadConfigFromYaml(testFile)
	want := map[string]string{
		"rome.ct.filippo.io/2024h2": "/etc/sunlight/rome2024h2.key",
		"rome.ct.filippo.io/2025h1": "/etc/sunlight/rome2025h1.key",
		"rome.ct.filippo.io/2025h2": "/etc/sunlight/rome2025h2.key",
	}

	if !reflect.DeepEqual(got, want) || err != nil {
		t.Errorf("got %q and error %q, wanted nil and not nil error", got, err)
	}
}

// Represent error cases.
var (
	errSecretIDNil    = errors.New("SecretId cannot be nil")
	errSecretNotFound = errors.New("secret not found")
)

// MockSecretsManagerAPI type mocks output and error responses returned from AWS Secrets Manager SDK.
type mockSecretsManagerAPI func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)

// Implements GetSecretValue function from Secret Manager for mock.
func (m mockSecretsManagerAPI) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return m(ctx, params, optFns...)
}

// Returns mock implementation of Secrets Manager SDK.
// Simulates fetching secrets based on provided seeds.
func newMockSecretsManagerAPI(seeds map[string]string) mockSecretsManagerAPI {
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

// Tests the FetchSecrets tests FetchSecrets function by using mock implementations of Secrets Manager SDK.
// Verifies that correct secrets are fetched or appropriate errors are returned.
func TestFetchSecrets(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		client mockSecretsManagerAPI
		seeds  map[string]string
		expect []string
		err    error
	}{
		{
			name:   "secret retrieved",
			client: newMockSecretsManagerAPI(map[string]string{"SECRET_1": "value1", "SECRET_2": "value2"}),
			seeds:  map[string]string{"KEY1": "SECRET_1", "KEY2": "SECRET_2"},
			expect: []string{"SECRET_1", "SECRET_2"},
			err:    nil,
		},
		{
			name:   "secret not found",
			client: newMockSecretsManagerAPI(map[string]string{}),
			seeds:  map[string]string{"KEY1": "SECRET_1"},
			expect: nil,
			err:    errSecretNotFound,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runTestFetchSecrets(t, tt.client, tt.seeds, tt.expect, tt.err)
		})
	}
}

// Helper function to TestFetchSecrets.
// Compares output of FetchSecrets with expected behavior.
func runTestFetchSecrets(t *testing.T, client mockSecretsManagerAPI, seeds map[string]string, expect []string, expectedErr error) {
	t.Helper()

	returnedKeys, err := config.FetchSecrets(seeds, client)

	if expectedErr != nil {
		if err == nil || !errors.Is(err, expectedErr) {
			log.Printf("Test failed: expected error %v, got %v", expectedErr, err)

			t.Fatalf("expected error %v, got %v", expectedErr, err)
		}
	} else {
		if err != nil {
			log.Printf("Test failed with unexpected error: %v", err)
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if len(expect) != len(returnedKeys) {
		log.Printf("Test failed: expected %v keys, got %v keys", len(expect), len(returnedKeys))
		t.Fatalf("expected %v keys, got %v keys", len(expect), len(returnedKeys))
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
			log.Printf("Test failed: expected key %s not found in returned keys", key)
			t.Errorf("expected key %s not found in returned keys", key)
		}
	}
}
