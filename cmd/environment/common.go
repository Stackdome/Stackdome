package environment

import (
	"fmt"
	"os"

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
