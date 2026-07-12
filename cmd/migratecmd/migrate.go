package migratecmd

import (
	"context"

	"github.com/Stackdome/stackdome/cmd/environment"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/spf13/cobra"
)

var log = logger.NewLogger()

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
		log.Fatalf("Unable to initialize environment: %s", err.Error())
	}
	db.Migrate(env.Environment().DBSession.New(context.Background()))
}
