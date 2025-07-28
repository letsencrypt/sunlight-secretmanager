package main

import (
	"fmt"
	"os"
	"syscall"
)

func writeFile(path string, content []byte, fsType int64) error {
	// We currently assume that the directory we are trying to write to already exists.
	file, err := os.OpenFile(
		path,
		// The combination of O_CREATE and O_EXCL means this operation will fail
		// if the file already exists.
		os.O_RDWR|os.O_CREATE|os.O_EXCL,
		// Setting nolint here because file permissions octal value isn't a magic number.
		//nolint: mnd
		0o400,
	)
	if err != nil {
		return fmt.Errorf("creating file at path %q: %w", path, err)
	}
	defer file.Close()

	var statfs syscall.Statfs_t
	err = syscall.Fstatfs(int(file.Fd()), &statfs)
	if err != nil {
		_ = os.Remove(file.Name())

		return fmt.Errorf("getting filesystem info at path %q: %w", path, err)
	}

	if statfs.Type != fsType {
		_ = os.Remove(file.Name())

		return fmt.Errorf("filesystem at path %q has type %v, but we require %v", path, statfs.Type, fsType)
	}

	_, err = file.Write(content)
	if err != nil {
		_ = os.Remove(file.Name())

		return fmt.Errorf("writing to file at path %q: %w", path, err)
	}

	return nil
}
