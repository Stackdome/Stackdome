package migratecmd

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/cmd/environment"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func NewMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Long:  "Run database migrations",
	}
	env := environment.LoadEnv()
	cmd.Run = func(cmd *cobra.Command, args []string) {
		runMigrate(env)
	}
	return cmd
}

func runMigrate(env environment.EnvImpl) {
	if err := env.InitDatabase(context.Background()); err != nil {
		glog.Exitf("Unable to initialize environment: %s", err.Error())
	}
	db.Migrate(env.Environment().DBSession.New(context.Background()))
}
