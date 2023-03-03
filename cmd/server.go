package cmd

import (
	"github.com/astaclinic/astafx/config"
	"github.com/astaclinic/astafx/graceful"
	"github.com/spf13/cobra"

	"github.com/laminatedio/dendrite/internal/app"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the API server",
	Long:  `Run the API server.`,
	Run: func(cmd *cobra.Command, args []string) {
		config.InitConfig(cfgFile)
		graceful.Run(app.New())
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
