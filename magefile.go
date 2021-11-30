//go:build mage
// +build mage

package main

import (
	"fmt"
	"io/ioutil"
	"os"
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
	protopath := filepath.Join(baseDir, "protobuf_files")
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

	err := sh.Run("go", "test", "-covermode=atomic", "-coverprofile=coverage.out", "./...")
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
