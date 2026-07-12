package environment

import (
	"fmt"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/logger"
)

var log = logger.NewLogger()

const (
	DEVELOPMENT_ENV = "DEVELOPMENT"
	PRODUCTION_ENV  = "PRODUCTION"
	TESTING_ENV     = "TESTING"
)

func GetEnvironmentStrFromEnv() string {
	val, ok := config.EnvStackdomeEnv.Lookup()
	if !ok || val == "" {
		log.Infof("Environment variable %q not specified, using default %q", config.EnvStackdomeEnv.Name, DEVELOPMENT_ENV)
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
