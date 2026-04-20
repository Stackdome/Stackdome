package environment

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/golang/glog"
)

const (
	EnvironmentStringKey = "STACKDOME_ENV"
	DEVELOPMENT_ENV      = "DEVELOPMENT"
	PRODUCTION_ENV       = "PRODUCTION"
	TESTING_ENV          = "TESTING"
	EnvironmentDefault   = DEVELOPMENT_ENV
)

func GetEnvironmentStrFromEnv() string {
	envStr, specified := os.LookupEnv(EnvironmentStringKey)
	if !specified || envStr == "" {
		glog.Infof("Environment variable %q not specified, using default %q", EnvironmentStringKey, EnvironmentDefault)
		envStr = EnvironmentDefault
	}
	return envStr
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
