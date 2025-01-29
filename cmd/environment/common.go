package environment

import (
	"fmt"
	"os"

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
