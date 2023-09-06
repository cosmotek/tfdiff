/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cosmotek/tfdiff/terraform"
	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
)

var enableVerboseLogging bool
var terraformWorkspaceName string
var ignoreAssetsGlobs []string

func parseIgnoreList() ([]string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get wd: %w", err)
	}

	tfdiffFilepath := fmt.Sprintf("%s/.tfdiff_ignore", wd)
	_, err = os.Stat(tfdiffFilepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}

		return nil, fmt.Errorf("failed to read .tfdiff_ignore: %w", err)
	}

	contents, err := os.ReadFile(tfdiffFilepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read .tfdiff_ignore: %w", err)
	}

	lines := strings.Split(string(contents), "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
		_, err = glob.Compile(lines[i])
		if err != nil {
			return nil, fmt.Errorf("failed to validate glob on line %d: %w", i, err)
		}
	}

	return lines, nil
}

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

	ignoreAssetsGlobs, err = parseIgnoreList()
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
