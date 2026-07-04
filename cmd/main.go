package main

import (
	"flag"

	"github.com/Stackdome/stackdome/cmd/migratecmd"
	"github.com/Stackdome/stackdome/cmd/servecmd"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:  "stackdome-server",
		Long: "stackdome-server",
	}

	// Always log to stderr by default
	if err := flag.Set("logtostderr", "true"); err != nil {
		glog.Infof("Unable to set logtostderr to true")
	}

	serveCmd := servecmd.NewServeCommand()
	migrateCmd := migratecmd.NewMigrateCommand()
	rootCmd.AddCommand(serveCmd, migrateCmd)

	if err := rootCmd.Execute(); err != nil {
		glog.Fatalf("error running command: %v", err)
	}
}
