/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var disablePrettyOutput bool
var outputFile string
var autoconfirm bool

// awsCmd represents the aws command
var gcpCmd = &cobra.Command{
	Use:   "gcp",
	Short: "run a diff against GCP cloud environment",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Not yet implemented, more coming soon...")
	},
}

func init() {
	gcpCmd.Flags().BoolVar(&disablePrettyOutput, "ugly", false, "disable pretty (colored) CLI output, cannot be used with 'output-file' flag")
	gcpCmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "specify the file to output diff results in CSV")
	gcpCmd.Flags().BoolVarP(&autoconfirm, "confirm", "y", false, "preconfirm execution parameters")

	rootCmd.AddCommand(gcpCmd)
}
