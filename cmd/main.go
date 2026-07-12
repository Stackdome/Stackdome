package main

import (
	"github.com/Stackdome/stackdome/cmd/migratecmd"
	"github.com/Stackdome/stackdome/cmd/servecmd"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/spf13/cobra"
)

var log = logger.NewLogger()

func main() {
	rootCmd := &cobra.Command{
		Use:  "stackdome-server",
		Long: "stackdome-server",
	}

	serveCmd := servecmd.NewServeCommand()
	migrateCmd := migratecmd.NewMigrateCommand()
	rootCmd.AddCommand(serveCmd, migrateCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("error running command: %v", err)
	}
}
