package main

import (
	"flag"

	"github.com/ashishmax31/soradev-api-server/cmd/migratecmd"
	"github.com/ashishmax31/soradev-api-server/cmd/servecmd"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:  "soradev-server",
		Long: "soradev-server",
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
