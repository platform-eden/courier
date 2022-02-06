//go:build mage
// +build mage

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/sh"
)

var (
	baseDir = getMageDir()
)

func getMageDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	return dir
}

// updates grpc boilerplate
func Proto() error {
	protopath := filepath.Join(baseDir, "proto")
	// get files in proto path
	files, err := ioutil.ReadDir(protopath)
	if err != nil {
		return fmt.Errorf("could not get files in %s: %s", baseDir, err)
	}

	// get the generated protobuf files for each proto file for go
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".proto") {

			err = sh.Run("protoc", "--proto_path="+protopath, "--go-grpc_out=.", file.Name())
			if err != nil {
				return fmt.Errorf("could not create go proto files: %s", err)
			}

			err = sh.Run("protoc", "--proto_path="+protopath, "--go_out=.", file.Name())
			if err != nil {
				return fmt.Errorf("could not create go proto files: %s", err)
			}
		}
	}

	return nil
}

// runs race tests
func Race() error {
	os.Chdir(baseDir)

	err := sh.Run("go", "test", "-race", "-covermode=atomic", "-coverprofile=coverage.out", "./pkg/...")
	if err != nil {
		return fmt.Errorf("failed unit test: %s", err)
	}

	return nil
}

// formats go code
func Fmt() error {
	os.Chdir(baseDir)

	err := sh.Run("go", "fmt", "./...")
	if err != nil {
		return fmt.Errorf("failed formatting: %s", err)
	}

	return nil
}

//creates mocks for pkg interfaces
func Mock() error {
	pkgDir := filepath.Join(baseDir, "pkg")
	err := filepath.Walk(pkgDir, mockWalkFunction)
	if err != nil {
		return fmt.Errorf("Mock: %w", err)
	}

	return nil
}

func mockWalkFunction(subDir string, info os.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("mockWalkFunction: %w", err)
	}

	pkgDir := filepath.Join(baseDir, "pkg")
	if subDir == pkgDir || path.Base(subDir) == "proto" {
		return nil
	}

	isDir, err := isDirectory(subDir)
	if err != nil {
		return fmt.Errorf("mockWalkFunction: %w", err)
	}

	if isDir {
		err = createMocks(subDir)
		if err != nil {
			return fmt.Errorf("mockWalkFunction: %w", err)
		}
	}

	return nil
}

func createMocks(subDir string) error {
	os.Chdir(subDir)

	err := sh.Run("mockery", "--all", "--case", "underscore")
	if err != nil {
		return fmt.Errorf("MakeMocks: %w", err)
	}

	return nil
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), nil
}
