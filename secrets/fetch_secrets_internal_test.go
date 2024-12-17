package secrets

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/smithy-go/middleware"
	"github.com/letsencrypt/sunlight-secretmanager/config"
)

// Represents error cases.
var (
	errSecretIDNil     = errors.New("SecretId cannot be nil")
	errSecretNotFound  = errors.New("secret not found")
	errFileCheckFailed = errors.New("filesystem check failed")
)

// mockSecretsManager API is a mock implementation of AWSSecretsManagerAPI interface.
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

// MockIsFilesystemFunc is a mock implementation of verifyFilesystemFunc.
type MockIsFilesystemFunc func(file *os.File, fs Filesystem) (bool, error)

// TestWriteToTmpfile defines test cases for the writeToTmpfile function using mock implementation of IsFilesystemFunc.
func TestWriteToTmpfile(t *testing.T) {
	testCases := []struct {
		name          string
		filename      config.FileType
		secret        []byte
		mockCheckFunc func(file *os.File, fs Filesystem) error
		expectedError error
	}{
		{
			name: "Successful",
			filename: config.FileType{
				Fullpath: "/etc/filename",
				Filename: "file.key",
			},
			secret: []byte{226, 151, 186},
			mockCheckFunc: func(_ *os.File, _ Filesystem) error {
				return nil
			},
			expectedError: nil,
		},
		{
			name: "Error",
			filename: config.FileType{
				Fullpath: "/etc/filename",
				Filename: "file.key",
			},
			secret: []byte{226, 151, 186},
			mockCheckFunc: func(_ *os.File, _ Filesystem) error {
				return errFileCheckFailed
			},
			expectedError: errInvalidTmpfs,
		},
	}

	t.Parallel()

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			runWriteToTmpfileTest(t, testcase)
		})
	}
}

// RunWriteToTmpfileTest is a helper function to TestWriteToTmpfile.
// It runs tests to verify that if a file is correctly mounted on tmpfs, the secrets is correctly written to the file.
func runWriteToTmpfileTest(t *testing.T, testcase struct {
	name          string
	filename      config.FileType
	secret        []byte
	mockCheckFunc func(file *os.File, fs Filesystem) error
	expectedError error
},
) {
	t.Helper()

	tempDir := t.TempDir()

	testFilename := config.FileType{
		Fullpath: filepath.Join(tempDir, "test.key"),
		Filename: "test.key",
	}

	verifyFilesystemFunc = testcase.mockCheckFunc

	result, err := writeToTmpfile(testcase.secret, testFilename, Filesystem(0x01021994))

	if !errors.Is(errors.Unwrap(err), errors.Unwrap(testcase.expectedError)) {
		t.Errorf("expected error %v got %v", testcase.expectedError, err)
	}

	if result != "" {
		if !strings.HasPrefix(result, tempDir) {
			t.Errorf("file not created in expected directory. Got %s, want prefix %s", result, tempDir)
		}

		if fileContent, readErr := os.ReadFile(result); readErr != nil {
			t.Errorf("failed to read result file: %v", readErr)
		} else if !bytes.Equal(fileContent, testcase.secret) {
			t.Errorf("file content mismatch")
		}
	}
}
