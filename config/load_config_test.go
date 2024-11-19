package config_test

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
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

type mockSecretsManagerAPI func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)

func (m mockSecretsManagerAPI) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return m(ctx, params, optFns...)
}

/*
func TestLoadSecrets_AllSecretsFound(t *testing.T) {
	client := mockSecretsManagerAPI(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
		 if params.SecretId == nil {
			  t.Fatal("expect SecretId to not be nil")
		 }
		 // Simulate successful retrieval for SECRET_1 and SECRET_2
		 if *params.SecretId == "SECRET_1" || *params.SecretId == "SECRET_2" {
			  return &secretsmanager.GetSecretValueOutput{Name: aws.String(*params.SecretId)}, nil
		 }
		 return nil, errors.New("secret not found")
	})

	seeds := map[string]string{
		 "KEY1": "SECRET_1",
		 "KEY2": "SECRET_2",
	}

	expectedKeys := []string{"KEY1", "KEY2"}

	returnedKeys, err := config.LoadSecrets(seeds, client)
	if err != nil {
		 t.Fatalf("unexpected error: %v", err)
	}

	if len(expectedKeys) != len(returnedKeys) {
		 t.Fatalf("expected %v keys, got %v keys", len(expectedKeys), len(returnedKeys))
	}

	for _, key := range expectedKeys {
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


func TestLoadSecrets_SomeSecretsFound(t *testing.T) {
	client := mockSecretsManagerAPI(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
		 if params.SecretId == nil {
			  t.Fatal("expect SecretId to not be nil")
		 }
		 // Simulate successful retrieval for SECRET_1 only
		 if *params.SecretId == "SECRET_1" {
			  return &secretsmanager.GetSecretValueOutput{Name: aws.String("SECRET_1")}, nil
		 }
		 return nil, errors.New("secret not found")
	})

	seeds := map[string]string{
		 "KEY1": "SECRET_1",
		 "KEY2": "SECRET_2",
	}

	expectedKeys := []string{"KEY1"}

	returnedKeys, err := config.LoadSecrets(seeds, client)
	if err == nil {
		 t.Fatal("expected error, got nil")
	}

	if len(expectedKeys) != len(returnedKeys) {
		 t.Fatalf("expected %v keys, got %v keys", len(expectedKeys), len(returnedKeys))
	}

	for _, key := range expectedKeys {
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
	

func TestLoadSecrets_NoSecretsFound(t *testing.T) {
	client := mockSecretsManagerAPI(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
		 return nil, errors.New("secret not found")
	})

	seeds := map[string]string{
		 "KEY1": "SECRET_1",
		 "KEY2": "SECRET_2",
	}

	returnedKeys, err := config.LoadSecrets(seeds, client)
	if err == nil {
		 t.Fatal("expected error, got nil")
	}

	if len(returnedKeys) != 0 {
		 t.Fatalf("expected 0 keys, got %v keys", len(returnedKeys))
	}
}
	*/




func TestLoadSecrets(t *testing.T) {
	cases := []struct {
		client func(t *testing.T) config.SecretsManagerAPI
		seeds  map[string]string
		expect []string
		err    error
	}{
		{
			client: func(t *testing.T) config.SecretsManagerAPI {
				return mockSecretsManagerAPI(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
					t.Helper()
					if params.SecretId == nil {
						t.Fatal("expect SecretId to not be nil")
					}
					if *params.SecretId == "SECRET_1" {
						return &secretsmanager.GetSecretValueOutput{Name: aws.String("SECRET_1")}, nil
					}
					return nil, errors.New("secret not found")
				})
			},
			seeds: map[string]string{
				"KEY1": "SECRET_1",
			},
			expect: []string{"KEY1"},
			err:    nil,
		},
		{
			client: func(t *testing.T) config.SecretsManagerAPI {
				return mockSecretsManagerAPI(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
					t.Helper()
					return nil, errors.New("secret not found")
				})
			},
			seeds: map[string]string{
				"KEY1": "SECRET_1",
			},
			expect: nil,
			err:    errors.New("secret not found"),
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			returnedKeys, err := config.LoadSecrets(tt.seeds, tt.client(t))

			if tt.err != nil {
				if err == nil || err.Error() != tt.err.Error() {
					t.Fatalf("expected error %v, got %v", tt.err, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if len(tt.expect) != len(returnedKeys) {
				t.Fatalf("expected %v keys, got %v keys", len(tt.expect), len(returnedKeys))
			}

			for _, key := range tt.expect {
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
		})
	}
}
	