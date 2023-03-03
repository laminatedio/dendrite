/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/astaclinic/astafx/info"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the version of the program",
	Run: func(cmd *cobra.Command, args []string) {
		info, err := info.GetInfo()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(info.Name)
		fmt.Println(info.Platform)
		fmt.Println(info.Runtime)
		fmt.Println(info.BuildCommit)
		fmt.Println(info.BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
