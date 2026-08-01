package environment

import (
	"fmt"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/logger"
)

var log = logger.NewLogger()

// Selectable values of STACKDOME_ENV. The test environment is not selectable
// here: it needs a session factory injected by the test bootstrap.
const (
	DEVELOPMENT_ENV = "DEVELOPMENT"
	PRODUCTION_ENV  = "PRODUCTION"
)

// envSpec holds everything that actually differs between environments. All
// initialization logic is shared; the spec parameterizes it.
type envSpec struct {
	// name is stamped on Env.Name and carried into workers/controllers.
	name string
	// logPrefix is prepended to logger and leadership-flag names.
	logPrefix string
	// managed means the environment loads its own config and builds its own
	// database session and email client. Tests get those injected instead.
	managed bool
}

var (
	developmentSpec = envSpec{name: config.EnvironmentDevelopment, managed: true}
	productionSpec  = envSpec{name: config.EnvironmentProduction, managed: true}
	testSpec        = envSpec{name: config.EnvironmentTest, logPrefix: "test-"}
)

var specsByRuntimeName = map[string]envSpec{
	DEVELOPMENT_ENV: developmentSpec,
	PRODUCTION_ENV:  productionSpec,
}

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
	spec, ok := specsByRuntimeName[envName]
	if !ok {
		panic(fmt.Sprintf("env: %s not defined", envName))
	}
	return newEnvironment(spec)
}
