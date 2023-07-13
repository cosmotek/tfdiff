/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/cosmotek/tfdiff/terraform"
	"github.com/spf13/cobra"
)

var enableVerboseLogging bool
var terraformWorkspaceName string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tfdiff",
	Short: "Generate reports for your migration from ClickOps to Terraform.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	var err error
	terraformWorkspaceName, err = terraform.GetWorkspace()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = rootCmd.Execute()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&enableVerboseLogging, "verbose", "v", false, "enable verbose logging")
}
