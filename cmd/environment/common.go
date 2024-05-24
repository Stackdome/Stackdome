package environment

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
)

const (
	EnvironmentStringKey = "SORADEV_ENV"
	DEVELOPMENT_ENV      = "DEVELOPMENT"
	PRODUCTION_ENV       = "PRODUCTION"
	TESTING_ENV          = "TESTING"
	EnvironmentDefault   = DEVELOPMENT_ENV
)

func setConfigDefaults(flags *pflag.FlagSet, defaults map[string]string) error {
	for name, value := range defaults {
		if err := flags.Set(name, value); err != nil {
			glog.Errorf("Error setting flag %s: %v", name, err)
			return err
		}
	}
	return nil
}

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
