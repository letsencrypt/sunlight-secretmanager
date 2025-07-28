package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type fakeSecretsManager struct{}

var _ SecretsManager = (*fakeSecretsManager)(nil)

func (sm *fakeSecretsManager) GetSecretValue(_ context.Context, params *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	if params.SecretId == nil || len(*params.SecretId) == 0 {
		return nil, errors.New("incomplete request")
	}

	switch *params.SecretId {
	case "missing":
		return nil, fmt.Errorf("secret %q not found", *params.SecretId)
	case "empty":
		return &secretsmanager.GetSecretValueOutput{ //nolint:exhaustruct
			Name: aws.String("empty"),
		}, nil
	case "real":
		return &secretsmanager.GetSecretValueOutput{ //nolint:exhaustruct
			Name:         aws.String("real"),
			SecretBinary: []byte("hello world"),
		}, nil
	default:
		return nil, fmt.Errorf("secret %q not recognized", *params.SecretId)
	}
}

func (sm *fakeSecretsManager) CreateSecret(_ context.Context, params *secretsmanager.CreateSecretInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error) {
	if params.Name == nil || len(*params.Name) == 0 || len(params.SecretBinary) == 0 {
		return nil, errors.New("incomplete request")
	}

	if params.SecretString != nil {
		return nil, errors.New("can't specify both SecretBinary and SecretString")
	}

	if *params.Name == "error" {
		return nil, fmt.Errorf("error 500 creating secret %q", *params.Name)
	}

	if len(params.SecretBinary) != 32 {
		return nil, fmt.Errorf("bad seed length: %d", len(params.SecretBinary))
	}

	return &secretsmanager.CreateSecretOutput{ //nolint:exhaustruct
		ARN:       aws.String(*params.Name + "-123456"),
		Name:      params.Name,
		VersionId: aws.String("919108f7-52d1-4320-9bac-f847db4148a8"),
	}, nil
}

func TestFetchSeed(t *testing.T) {
	t.Parallel()

	testSM := &fakeSecretsManager{}

	for _, tc := range []struct {
		name    string
		secret  string
		want    []byte
		wantErr string
	}{
		{
			name:    "bad request",
			secret:  "",
			want:    nil,
			wantErr: "incomplete request",
		},
		{
			name:    "missing secret",
			secret:  "missing",
			want:    nil,
			wantErr: "retrieving secret",
		},
		{
			name:    "empty",
			secret:  "empty",
			want:    []byte{},
			wantErr: "",
		},
		{
			name:    "happy path",
			secret:  "real",
			want:    []byte("hello world"),
			wantErr: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := fetchSeed(t.Context(), testSM, tc.secret)
			if tc.wantErr != "" { //nolint:nestif
				if err == nil {
					t.Errorf("fetchSeed(%q) = %#v, but want error %q", tc.secret, got, tc.wantErr)
				} else if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("fetchSeed(%q) = %#v, but want error %q", tc.secret, err, tc.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("fetchSeed(%q) = %#v, but want no error", tc.secret, err)
				} else if !bytes.Equal(got, tc.want) {
					t.Errorf("fetchSeed(%q) = %#v, want %#v", tc.secret, got, tc.want)
				}
			}
		})
	}
}

func TestCreateSeed(t *testing.T) {
	t.Parallel()

	testSM := &fakeSecretsManager{}

	for _, tc := range []struct {
		name    string
		id      string
		wantErr string
	}{
		{
			name:    "bad request",
			id:      "",
			wantErr: "incomplete request",
		},
		{
			name:    "error",
			id:      "error",
			wantErr: "creating secret \"error\": error 500 creating secret",
		},
		{
			name:    "happy path",
			id:      "happy",
			wantErr: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := createSeed(t.Context(), testSM, tc.id)
			if tc.wantErr != "" {
				if err == nil {
					t.Errorf("createSeed(%q) = %#v, but want error %q", tc.id, got, tc.wantErr)
				} else if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("createSeed(%q) = %#v, but want error %q", tc.id, err, tc.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("createSeed(%q) = %#v, but want no error", tc.id, err)
				}
			}
		})
	}
}
