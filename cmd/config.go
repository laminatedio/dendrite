package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	astafxConfig "github.com/astaclinic/astafx/config"
	"github.com/laminatedio/dendrite/internal/pkg/config"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Validate and display the config read",
	Run: func(cmd *cobra.Command, args []string) {
		astafxConfig.InitConfig(cfgFile)
		if err := config.PrintConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "%s\t[FATAL]\tfail to output config: %v\n", time.Now().Format(time.RFC3339), err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
