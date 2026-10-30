/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/cosmotek/tfdiff/scanner"
	"github.com/cosmotek/tfdiff/scanner/azure"
	"github.com/cosmotek/tfdiff/terraform"
	"github.com/gobwas/glob"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var azureSubscriptions []string
var excludeAzureResourceTypes []string
var azureOutputFile string
var azureStitchDir string

// azureCmd represents the azure command
var azureCmd = &cobra.Command{
	Use:   "azure",
	Short: "run a diff against Azure cloud environment",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.WarnLevel)
		if enableVerboseLogging {
			logger = logger.Level(zerolog.DebugLevel)
		}

		subscriptionScope := "all accessible"
		if len(azureSubscriptions) > 0 {
			subscriptionScope = strings.Join(azureSubscriptions, ",")
		}

		if azureStitchDir != "" {
			fmt.Printf("tfdiff starting: stitch_dir=%s azure_subscriptions=%s\n", azureStitchDir, subscriptionScope)
		} else {
			fmt.Printf("tfdiff starting: terraform_workspace=%s azure_subscriptions=%s\n", terraformWorkspaceName, subscriptionScope)
		}
		if len(azureSubscriptions) == 0 {
			fmt.Println("warning: scanning all accessible subscriptions may take a while")
		}
		time.Sleep(time.Second * 3) // sleep to allow cancellation

		azureScanner, err := azure.New(cmd.Context(), logger, azure.Config{
			Subscriptions:        azureSubscriptions,
			ExcludeResourceTypes: excludeAzureResourceTypes,
			MaxConcurrency:       1,
		})
		if err != nil {
			return err
		}

		s := spinner.New(spinner.CharSets[11], 165*time.Millisecond)
		s.Suffix = "\n"
		s.Reverse() // change the direction of the spinner
		startTime := time.Now()

		s.Prefix = "scanning azure environment for assets "
		s.Start()
		defer s.Stop()

		assetList, err := azureScanner.RunScan(cmd.Context())
		if err != nil {
			return err
		}

		// In stitch mode state is read from pre-pulled files covering many
		// projects at once; otherwise a single state is pulled from the CWD.
		var stitch *stitchStates
		var pullState terraform.PullStateOutput
		if azureStitchDir != "" {
			s.Prefix = "loading terraform states "
			s.Restart()

			stitch, err = loadStitchStates(azureStitchDir)
			if err != nil {
				return err
			}
		} else {
			s.Prefix = "reading terraform state "
			s.Restart()

			pullState, err = terraform.PullState()
			if err != nil {
				return err
			}
		}

		s.Prefix = "computing diff "
		s.Restart()

		for _, globStr := range ignoreAssetsGlobs {
			globC, err := glob.Compile(globStr)
			if err != nil {
				return fmt.Errorf("failed to compile glob '%s' from ignorelist: %w", globStr, err)
			}

			assetList = lo.Filter(assetList, func(asset scanner.Asset, _ int) bool {
				return !globC.Match(asset.Identifier)
			})
		}

		var managedResources scanner.AssetList
		if stitch != nil {
			managedResources = stitch.managed(assetList)
		} else {
			managedResources = lo.Filter(assetList, func(asset scanner.Asset, _ int) bool {
				for _, tfrsc := range pullState.Resources {
					for _, inst := range tfrsc.Instances {
						// azurerm stores the full ARM resource ID in the id
						// attribute; Azure resource IDs are case-insensitive.
						if strings.EqualFold(asset.Identifier, inst.Attributes.ID) {
							return true
						}
					}
				}

				return false
			})
		}

		unmanagedResources := lo.Filter(assetList, func(asset scanner.Asset, _ int) bool {
			for _, rsrc := range managedResources {
				if asset.Identifier == rsrc.Identifier {
					return false
				}
			}

			return true
		})

		regions := lo.GroupBy(unmanagedResources, func(asset scanner.Asset) string {
			return asset.Region
		})

		s.Stop()
		fmt.Printf("tfdiff completed in %s.\n", time.Since(startTime).String())

		if azureOutputFile != "" {
			csv, err := scanner.AssetList(unmanagedResources).ToCSV()
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(azureOutputFile, []byte(csv), os.ModePerm)
			if err != nil {
				return err
			}

			fmt.Printf("report csv written to: %s\n", azureOutputFile)
		}

		fmt.Println("\nfinal report:")
		fmt.Printf("managed (%d/%d - %f%%)\n", len(managedResources), len(assetList), (float64(len(managedResources)) / float64(len(assetList)) * 100))
		fmt.Printf("unmanaged (%d/%d - %f%%)\n", len(unmanagedResources), len(assetList), (float64(len(unmanagedResources)) / float64(len(assetList)) * 100))

		fmt.Println("\nunmanaged asset breakdown:")
		for region, regionalAssets := range regions {
			fmt.Printf("\tregion %s (%d/%d - %f%%):\n", region, len(regionalAssets), len(unmanagedResources), (float64(len(regionalAssets)) / float64(len(unmanagedResources)) * 100))
			types := lo.GroupBy(regionalAssets, func(asset scanner.Asset) string {
				return asset.ResourceType
			})

			entries := scanner.AssetCounterEntries(lo.Entries(types))
			sort.Sort(entries)

			for _, entry := range entries {
				fmt.Printf("\t\t%s %d\n", entry.Key, len(entry.Value))
			}
		}

		if stitch != nil {
			stitch.report(managedResources)
		}

		return nil
	},
}

func init() {
	azureCmd.Flags().StringSliceVarP(&azureSubscriptions, "subscriptions", "s", []string{}, "specify which azure subscription ids to scan (defaults to all accessible)")
	azureCmd.Flags().StringSliceVarP(&excludeAzureResourceTypes, "exclude-resources", "e", []string{}, "specify which azure resource types to exclude from scans")
	azureCmd.Flags().StringVarP(&azureOutputFile, "output-file", "o", "", "specify the file to output diff results in CSV")
	azureCmd.Flags().StringVar(&azureStitchDir, "stitch", "", "diff against every *.json state file in this directory at once, reporting per-project coverage (filename is the project name)")

	rootCmd.AddCommand(azureCmd)
}
