package servecmd

import (
	"context"

	"github.com/Stackdome/stackdome/cmd/environment"
	"github.com/Stackdome/stackdome/cmd/server"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/spf13/cobra"
)

var log = logger.NewLogger()

func NewServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "start the api server",
		Long:  "start the api server",
	}
	env := environment.LoadEnv()
	cmd.Run = func(cmd *cobra.Command, args []string) {
		runServe(env)
	}
	return cmd
}

func runServe(env environment.EnvImpl) {
	if err := env.Init(context.Background()); err != nil {
		log.Fatalf("Unable to initialize environment: %s", err.Error())
	}
	// Run the servers
	go func() {
		apiserver := server.NewAPIServer(env)
		apiserver.Start()
	}()

	select {}
}
