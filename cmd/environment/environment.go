package environment

import (
	"context"

	"github.com/ashishmax31/soradev-api-server/config"
	"github.com/ashishmax31/soradev-api-server/pkg/db"
	"github.com/ashishmax31/soradev-api-server/pkg/services"
	"github.com/spf13/pflag"
)

type EnvImpl interface {
	Init(context.Context) error
	InitDatabase(context.Context) error
	AddFlags(flags *pflag.FlagSet) error
	Environment() *Env
}

type Env struct {
	Name      string
	Services  Services
	DBSession db.SessionFactory
	Config    *config.ApplicationConfig
}

type Database struct {
	SessionFactory db.SessionFactory
}

type Services struct {
	UserService services.UserService
}
