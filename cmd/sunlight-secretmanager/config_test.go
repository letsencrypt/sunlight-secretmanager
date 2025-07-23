package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		input   string
		want    *config
		wantErr string
	}{
		{
			name:    "invalid path",
			input:   "",
			want:    nil,
			wantErr: "reading config file",
		},
		{
			name:    "invalid yaml",
			input:   "invalid.yaml",
			want:    nil,
			wantErr: "parsing config file",
		},
		{
			name:    "empty",
			input:   "empty.yaml",
			want:    nil,
			wantErr: "no logs found in config file", // an empty file is valid yaml
		},
		{
			name:    "no logs",
			input:   "nologs.yaml",
			want:    nil,
			wantErr: "no logs found in config file",
		},
		{
			name:    "incomplete",
			input:   "missingkeys.yaml",
			want:    nil,
			wantErr: "incomplete config for log",
		},
		{
			name:  "happy path",
			input: "happy.yaml",
			want: &config{
				Logs: []logConfig{
					{
						Name:      "test.tld/shard1",
						Inception: "2024-08-07",
						Secret:    "/path/to/shard1.seed",
					},
					{
						Name:      "test.tld/shard2",
						Inception: "2024-08-07",
						Secret:    "/path/to/shard2.seed",
					},
				},
			},
			wantErr: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := loadConfig(filepath.Join("testdata", tc.input))

			if tc.wantErr != "" { //nolint:nestif
				if err == nil {
					t.Errorf("loadConfig() = %#v, but want %q", got, tc.wantErr)
				} else if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("loadConfig() = %q, but want %q", err, tc.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("loadConfig() = %q, but want %#v", err, tc.want)
				} else if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("loadConfig() = %#v, but want %#v", got, tc.want)
				}
			}
		})
	}
}
