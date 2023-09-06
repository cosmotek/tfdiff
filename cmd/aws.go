/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/briandowns/spinner"
	"github.com/cosmotek/tfdiff/scanner"
	"github.com/cosmotek/tfdiff/scanner/aws"
	"github.com/cosmotek/tfdiff/terraform"
	"github.com/gobwas/glob"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var awsRegions []string
var defaultAwsRegions = aws.Regions
var resourceExplorerRegion string
var outputFile string

// awsCmd represents the aws command
var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "run a diff against AWS cloud environment",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.WarnLevel)
		if enableVerboseLogging {
			logger = logger.Level(zerolog.DebugLevel)
		}

		awsProfile := os.Getenv("AWS_PROFILE")
		if awsProfile == "" {
			return errors.New("$AWS_PROFILE must not be empty")
		}

		fmt.Printf("tfdiff starting: terraform_workspace=%s aws_profile=%s\n", terraformWorkspaceName, awsProfile)
		if len(awsRegions) == len(defaultAwsRegions) {
			fmt.Println("warning: scanning all regions may take a while")
		}
		time.Sleep(time.Second * 3) // sleep to allow cancellation

		awsScanner, err := aws.New(logger, aws.Config{
			ScanRegions:               awsRegions,
			ResourceExplorerAWSRegion: resourceExplorerRegion,
			MaxConcurrency:            1,
		})
		if err != nil {
			return err
		}

		s := spinner.New(spinner.CharSets[11], 165*time.Millisecond)
		s.Suffix = "\n"
		s.Reverse() // change the direction of the spinner
		startTime := time.Now()

		s.Prefix = "scanning aws environment for assets "
		s.Start()
		defer s.Stop()

		assetList, err := awsScanner.RunScan()
		if err != nil {
			return err
		}

		s.Prefix = "reading terraform state "
		s.Restart()

		pullState, err := terraform.PullState()
		if err != nil {
			return err
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

		managedResources := lo.Filter(assetList, func(asset scanner.Asset, _ int) bool {
			for _, tfrsc := range pullState.Resources {
				for _, inst := range tfrsc.Instances {
					if asset.Identifier == inst.Attributes.Arn {
						return true
					}
				}
			}

			return false
		})

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

		if outputFile != "" {
			csv, err := scanner.AssetList(unmanagedResources).ToCSV()
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(outputFile, []byte(csv), os.ModePerm)
			if err != nil {
				return err
			}

			fmt.Printf("report csv written to: %s\n", outputFile)
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

		return nil
	},
}

func init() {
	awsCmd.Flags().StringVarP(&resourceExplorerRegion, "resource-explorer-region", "e", "us-east-1", "specify which aws region the resource explorer instance is active in")
	awsCmd.Flags().StringSliceVarP(&awsRegions, "regions", "r", defaultAwsRegions, "specify which aws regions to scan for inventory")
	awsCmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "specify the file to output diff results in CSV")

	rootCmd.AddCommand(awsCmd)
}
