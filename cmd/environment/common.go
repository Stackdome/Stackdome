package environment

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ashishmax31/stackdome-api-server/config"
	"github.com/golang/glog"
)

const (
	DEVELOPMENT_ENV = "DEVELOPMENT"
	PRODUCTION_ENV  = "PRODUCTION"
	TESTING_ENV     = "TESTING"
)

func GetEnvironmentStrFromEnv() string {
	val, ok := config.EnvStackdomeEnv.Lookup()
	if !ok || val == "" {
		glog.Infof("Environment variable %q not specified, using default %q", config.EnvStackdomeEnv.Name, DEVELOPMENT_ENV)
		return DEVELOPMENT_ENV
	}
	return val
}

func LoadEnv() EnvImpl {
	envName := GetEnvironmentStrFromEnv()
	switch envName {
	case DEVELOPMENT_ENV:
		return NewDevelopmentEnvironment()
	default:
		panic(fmt.Sprintf("env: %s not defined", envName))
	}
}

func findGoModDir() (string, error) {
	// Get the caller's file path
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("unable to get caller information")
	}

	// Start from the directory containing this file
	dir := filepath.Dir(filename)

	// Walk up looking for go.mod
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}
