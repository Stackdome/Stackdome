package servecmd

import (
	"context"

	"github.com/ashishmax31/soradev-api-server/cmd/environment"
	"github.com/ashishmax31/soradev-api-server/cmd/server"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func NewServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "start the api server",
		Long:  "start the api server",
	}
	env := environment.LoadEnv()
	err := env.AddFlags(cmd.PersistentFlags())
	if err != nil {
		glog.Fatalf("Unable to add environment flags to serve command: %s", err.Error())
	}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		runServe(env)
	}
	return cmd
}

func runServe(env environment.EnvImpl) {
	if err := env.Init(context.Background()); err != nil {
		glog.Exitf("Unable to initialize environment: %s", err.Error())
	}
	// Run the servers
	go func() {
		apiserver := server.NewAPIServer(env)
		apiserver.Start()
	}()
	
	select {}
}
