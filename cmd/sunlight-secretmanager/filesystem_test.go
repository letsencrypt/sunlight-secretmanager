package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFile(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	// Do some setup for the "already exists" test.
	f, err := os.Create(filepath.Join(tempDir, "already-exists"))
	if err != nil {
		t.Fatalf("failed to create test setup file: %s", err)
	}
	err = f.Close()
	if err != nil {
		t.Fatalf("failed to close test setup file: %s", err)
	}

	for _, tc := range []struct {
		name    string
		path    string
		fsType  int64
		wantErr string
	}{
		{
			name:    "already exists",
			path:    filepath.Join(tempDir, "already-exists"),
			fsType:  1,
			wantErr: "creating file at path",
		},
		{
			name:    "wrong fstype",
			path:    filepath.Join(tempDir, "wrong-fstype"),
			fsType:  1,
			wantErr: "filesystem at path",
		},
		{
			name:    "happy path",
			path:    filepath.Join(tempDir, "happy"),
			fsType:  61267, // The statfs.Type for a normal unix filesystem
			wantErr: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := writeFile(tc.path, []byte("hello world"), tc.fsType)

			if tc.wantErr != "" { //nolint:nestif
				if err == nil {
					t.Errorf("writeFile() = succeeded, but want error %q", tc.wantErr)
				} else if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("writeFile() = %#v, but want %q", err, tc.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("writeFile() = %#v, but want success", err)
				}

				got, err := os.ReadFile(tc.path)
				if err != nil {
					t.Fatalf("failed to re-read file: %s", err)
				}

				if string(got) != "hello world" {
					t.Errorf("written file contains %q, but want %q", string(got), "hello world")
				}
			}
		})
	}
}
