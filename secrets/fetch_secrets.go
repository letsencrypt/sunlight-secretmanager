package secrets

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/letsencrypt/sunlight-secretmanager/config"
)

// ErrInvalidTmpfs represents error case in which check that file is on a tmpfs fails.
var errInvalidTmpfs = errors.New("invalid tmpfs mount: filesystem check failed")

// AWSSecretsManagerAPI defines the interface for the AWS Secrets Manager operations required by FetchSecretsHelper.
type AWSSecretsManagerAPI interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type Secrets struct {
	svc AWSSecretsManagerAPI
}

// Filesystem represents tmpfs const.
type Filesystem int64

func New(cfg aws.Config) *Secrets {
	return &Secrets{
		svc: secretsmanager.NewFromConfig(cfg),
	}
}

// FetchSecrets uses Config Profile to initialize AWS SDK configuration.
// Calls FetchSecretsHelper and passes it configured AWS Secrets Manager client.
func (s *Secrets) FetchSecrets(ctx context.Context, seeds map[string]string, fileNamesMap map[string]config.FileType, fsConst Filesystem) (map[string][]byte, error) {
	returnedKeys := make(map[string][]byte)

	for _, seedValue := range seeds {
		input := &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(seedValue),
			VersionStage: aws.String("AWSCURRENT"),
			VersionId:    nil,
		}

		result, err := s.svc.GetSecretValue(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve secret for %v: %w", *input.SecretId, err)
		}

		res := *result
		secretName := res.Name
		secretValue := res.SecretBinary

		returnedKeys[*secretName] = secretValue

		file, err := writeToTmpfile(secretValue, fileNamesMap[*secretName], fsConst)
		if file == "" || err != nil {
			return nil, fmt.Errorf("couldn't write secret to file %v with error %w because filesystem const is %v", file, err, fsConst)
		}
	}

	return returnedKeys, nil
}

// WritetoTmpFile opens a file with restrictive user-only-read permissions and writes content to the file if it is on tmpfs.
func writeToTmpfile(secret []byte, fileNames config.FileType, fsConst Filesystem) (string, error) {
	tmpFile, err := os.OpenFile(
		filepath.Join(fileNames.Fullpath, fileNames.Filename),
		os.O_RDWR|os.O_CREATE|os.O_EXCL,
		//nolint: mnd
		0o604,
	)
	if err != nil {
		return "", fmt.Errorf("didn't create tmpfile called %v with error %w", tmpFile, err)
	}

	defer tmpFile.Close()

	isLinux, err := isFilesystemFunc(tmpFile, fsConst)
	if err != nil || !isLinux {
		os.Remove(tmpFile.Name())

		return "", errInvalidTmpfs
	}

	if _, err := tmpFile.Write(secret); err != nil {
		os.Remove(tmpFile.Name())

		return "", fmt.Errorf("couldn't write to tmpfile with error %w", err)
	}

	return tmpFile.Name(), nil
}

// isFilesystemFunc verifies that file is on tmpfs.
var isFilesystemFunc = func(file *os.File, fsConst Filesystem) (bool, error) {
	var statfs syscall.Statfs_t
	err := syscall.Fstatfs(int(file.Fd()), &statfs)
	if err != nil {
		return false, errInvalidTmpfs
	}

	return Filesystem(statfs.Type) == fsConst, nil
}
