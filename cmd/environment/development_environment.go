package environment

import (
	"context"

	"github.com/ashishmax31/soradev-api-server/config"
	"github.com/ashishmax31/soradev-api-server/pkg/db"
	"github.com/ashishmax31/soradev-api-server/pkg/logger"
	"github.com/ashishmax31/soradev-api-server/pkg/services"
	"github.com/spf13/pflag"
)

type developmentEnvironment struct {
	*Env
}

func NewDevelopmentEnvironment() EnvImpl {
	return &developmentEnvironment{
		Env: &Env{
			Name:   "development",
			Config: config.NewApplicationConfig(),
		},
	}
}

func (d *developmentEnvironment) Environment() *Env {
	return d.Env
}

func (d *developmentEnvironment) AddFlags(flags *pflag.FlagSet) error {
	d.Config.AddFlags(flags)
	return setConfigDefaults(flags, d.defaultFlags())
}

func (d *developmentEnvironment) Init(ctx context.Context) error {
	d.Config.ReadEnvironmentVariables()
	if err := d.Config.ReadConfigFiles(); err != nil {
		return err
	}
	logger := logger.NewLogger(ctx)
	d.DBSession = db.NewSessionFactory(d.Config.Database)
	d.Services = d.loadSevices(logger)
	return nil
}

func (d *developmentEnvironment) loadSevices(logger logger.Logger) Services {
	return Services{
		UserService: services.NewUserService(services.UserServiceSpec{
			SessionFactory: d.DBSession,
			Logger:         logger,
			JwtSecretKey:   d.Config.Server.JwtSecret,
		}),
	}
}

func (d *developmentEnvironment) defaultFlags() map[string]string {
	return map[string]string{
		"v":                      "10",
		"api-server-hostname":    "localhost",
		"api-server-bindaddress": "localhost:8000",
	}
}
