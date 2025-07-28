package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type fakeSecretsManager struct{}

var _ SecretsManager = (*fakeSecretsManager)(nil)

func (sm *fakeSecretsManager) GetSecretValue(_ context.Context, params *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
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
