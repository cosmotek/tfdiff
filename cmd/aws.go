/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var awsRegions []string
var defaultAwsRegions = []string{"us-east-1", "us-east-2"}

var disablePrettyOutput bool
var outputFile string
var autoconfirm bool

// awsCmd represents the aws command
var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "run a diff against AWS cloud environment",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aws called")
	},
}

func init() {
	awsCmd.Flags().StringArrayVar(&awsRegions, "regions", defaultAwsRegions, "specify which aws regions to scan for inventory")
	awsCmd.Flags().BoolVar(&disablePrettyOutput, "ugly", false, "disable pretty (colored) CLI output, cannot be used with 'output-file' flag")
	awsCmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "specify the file to output diff results in CSV")
	awsCmd.Flags().BoolVarP(&autoconfirm, "confirm", "y", false, "preconfirm execution parameters")

	rootCmd.AddCommand(awsCmd)
}
