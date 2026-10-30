/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cosmotek/tfdiff/scanner"
	"github.com/cosmotek/tfdiff/terraform"
	"github.com/samber/lo"
)

// stitchStates holds the managed-resource identifiers declared across several
// projects' Terraform state files, keyed for diffing against a single scan.
type stitchStates struct {
	// coverage maps a normalized (lowercased) resource id to the projects
	// whose state declares it. More than one project is a conflict.
	coverage map[string][]string
	// declared counts the unique ids each project declares, used to derive
	// how much of a project's state was not found in the live scan.
	declared map[string]int
	// projects lists every state file loaded, in sorted order.
	projects []string
}

// loadStitchStates reads every *.json file in dir as a Terraform state export,
// using each file's name (without extension) as the project name.
func loadStitchStates(dir string) (*stitchStates, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read stitch dir %s: %w", dir, err)
	}

	ss := &stitchStates{
		coverage: map[string][]string{},
		declared: map[string]int{},
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}

		project := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		state, err := terraform.LoadState(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}

		seen := map[string]bool{}
		for _, resource := range state.Resources {
			if resource.Mode != "managed" {
				continue // data sources do not manage anything
			}

			for _, instance := range resource.Instances {
				id := instance.Attributes.ID
				if id == "" {
					continue
				}

				normalized := strings.ToLower(id)
				if seen[normalized] {
					continue
				}

				seen[normalized] = true
				ss.coverage[normalized] = append(ss.coverage[normalized], project)
			}
		}

		ss.declared[project] = len(seen)
		ss.projects = append(ss.projects, project)
	}

	if len(ss.projects) == 0 {
		return nil, fmt.Errorf("no .json state files found in %s", dir)
	}

	sort.Strings(ss.projects)
	return ss, nil
}

// managed returns the subset of assets whose identifier is declared by any
// loaded state.
func (ss *stitchStates) managed(assets scanner.AssetList) scanner.AssetList {
	return lo.Filter(assets, func(asset scanner.Asset, _ int) bool {
		_, ok := ss.coverage[strings.ToLower(asset.Identifier)]
		return ok
	})
}

// report prints per-project coverage, stale state entries, and conflicts.
func (ss *stitchStates) report(managed scanner.AssetList) {
	live := map[string]int{}
	for _, asset := range managed {
		for _, project := range ss.coverage[strings.ToLower(asset.Identifier)] {
			live[project]++
		}
	}

	fmt.Println("\ncoverage by project (live resources managed):")
	for _, project := range sortByCountDesc(ss.projects, live) {
		fmt.Printf("\t%s %d\n", project, live[project])
	}

	stale := map[string]int{}
	staleProjects := []string{}
	for _, project := range ss.projects {
		if n := ss.declared[project] - live[project]; n > 0 {
			stale[project] = n
			staleProjects = append(staleProjects, project)
		}
	}

	if len(staleProjects) > 0 {
		fmt.Println("\nstale state (declared in state, not found live):")
		for _, project := range sortByCountDesc(staleProjects, stale) {
			fmt.Printf("\t%s %d\n", project, stale[project])
		}
	}

	conflicts := [][2]string{}
	for id, projects := range ss.coverage {
		if len(projects) > 1 {
			sorted := append([]string{}, projects...)
			sort.Strings(sorted)
			conflicts = append(conflicts, [2]string{id, strings.Join(sorted, ", ")})
		}
	}

	if len(conflicts) > 0 {
		sort.Slice(conflicts, func(i, j int) bool { return conflicts[i][0] < conflicts[j][0] })
		fmt.Println("\nconflicts (resource claimed by >1 project):")
		for _, conflict := range conflicts {
			fmt.Printf("\t%s\t%s\n", conflict[0], conflict[1])
		}
	}
}

func sortByCountDesc(keys []string, counts map[string]int) []string {
	sorted := append([]string{}, keys...)
	sort.Slice(sorted, func(i, j int) bool {
		if counts[sorted[i]] != counts[sorted[j]] {
			return counts[sorted[i]] > counts[sorted[j]]
		}

		return sorted[i] < sorted[j]
	})

	return sorted
}
