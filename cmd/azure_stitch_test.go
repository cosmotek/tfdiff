package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cosmotek/tfdiff/scanner"
)

const vmID = "/subscriptions/S/resourceGroups/RG/providers/Microsoft.Compute/virtualMachines/VM1"
const diskID = "/subscriptions/S/resourceGroups/RG/providers/Microsoft.Compute/disks/DISK1"

func writeState(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestLoadStitchStates(t *testing.T) {
	dir := t.TempDir()

	// network manages the VM; a data source must be ignored.
	writeState(t, dir, "network.json", `{"resources":[
		{"mode":"managed","type":"azurerm_linux_virtual_machine","name":"vm","instances":[{"attributes":{"id":"`+vmID+`"}}]},
		{"mode":"data","type":"azurerm_resource_group","name":"rg","instances":[{"attributes":{"id":"/subscriptions/S/resourceGroups/IGNORED"}}]}
	]}`)
	// apps manages a disk, and also claims the VM -> a conflict.
	writeState(t, dir, "apps.json", `{"resources":[
		{"mode":"managed","type":"azurerm_managed_disk","name":"d","instances":[{"attributes":{"id":"`+diskID+`"}}]},
		{"mode":"managed","type":"azurerm_linux_virtual_machine","name":"vm","instances":[{"attributes":{"id":"`+vmID+`"}}]}
	]}`)
	// a non-json file must be skipped.
	writeState(t, dir, "notes.txt", "ignore me")

	ss, err := loadStitchStates(dir)
	if err != nil {
		t.Fatalf("loadStitchStates: %v", err)
	}

	if got := len(ss.projects); got != 2 {
		t.Fatalf("projects = %d, want 2 (%v)", got, ss.projects)
	}
	if ss.projects[0] != "apps" || ss.projects[1] != "network" {
		t.Fatalf("projects not sorted: %v", ss.projects)
	}

	// the data source must not count toward declared ids.
	if ss.declared["network"] != 1 {
		t.Fatalf("network declared = %d, want 1 (data source should be skipped)", ss.declared["network"])
	}

	// the VM is claimed by both projects -> conflict. Coverage keys are
	// normalized to lowercase, so look it up that way.
	if projects := ss.coverage[strings.ToLower(vmID)]; len(projects) != 2 {
		t.Fatalf("vm coverage = %v, want 2 projects", projects)
	}
}

func TestManagedMatchesCaseInsensitively(t *testing.T) {
	dir := t.TempDir()
	writeState(t, dir, "network.json", `{"resources":[
		{"mode":"managed","type":"azurerm_linux_virtual_machine","name":"vm","instances":[{"attributes":{"id":"`+vmID+`"}}]}
	]}`)

	ss, err := loadStitchStates(dir)
	if err != nil {
		t.Fatalf("loadStitchStates: %v", err)
	}

	assets := scanner.AssetList{
		// Resource Graph often returns a different casing than state.
		{Identifier: "/subscriptions/s/resourcegroups/rg/providers/microsoft.compute/virtualmachines/vm1"},
		{Identifier: diskID}, // unmanaged
	}

	managed := ss.managed(assets)
	if len(managed) != 1 {
		t.Fatalf("managed = %d, want 1", len(managed))
	}
	if managed[0].Identifier != assets[0].Identifier {
		t.Fatalf("managed matched wrong asset: %s", managed[0].Identifier)
	}
}

func TestLoadStitchStatesEmptyDir(t *testing.T) {
	if _, err := loadStitchStates(t.TempDir()); err == nil {
		t.Fatal("expected error for a directory with no state files")
	}
}

func TestReportCoverageAndStale(t *testing.T) {
	const ghostID = "/subscriptions/S/resourceGroups/RG/providers/Microsoft.Compute/disks/GONE"
	dir := t.TempDir()
	// network declares vm, disk, and a ghost that no longer exists live.
	writeState(t, dir, "network.json", `{"resources":[
		{"mode":"managed","type":"t","name":"vm","instances":[{"attributes":{"id":"`+vmID+`"}}]},
		{"mode":"managed","type":"t","name":"d","instances":[{"attributes":{"id":"`+diskID+`"}}]},
		{"mode":"managed","type":"t","name":"g","instances":[{"attributes":{"id":"`+ghostID+`"}}]}
	]}`)
	// apps also claims the disk -> conflict.
	writeState(t, dir, "apps.json", `{"resources":[
		{"mode":"managed","type":"t","name":"d","instances":[{"attributes":{"id":"`+diskID+`"}}]}
	]}`)

	ss, err := loadStitchStates(dir)
	if err != nil {
		t.Fatalf("loadStitchStates: %v", err)
	}

	managed := scanner.AssetList{{Identifier: vmID}, {Identifier: diskID}}
	out := captureStdout(t, func() { ss.report(managed) })

	for _, want := range []string{
		"network 2", // vm + disk live
		"apps 1",    // disk live
		"network 1", // ghost declared but not found live
	} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing %q in:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "conflicts") || !strings.Contains(out, strings.ToLower(diskID)) {
		t.Errorf("report missing disk conflict in:\n%s", out)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = orig

	buf := make([]byte, 0, 4096)
	chunk := make([]byte, 1024)
	for {
		n, readErr := r.Read(chunk)
		buf = append(buf, chunk[:n]...)
		if readErr != nil {
			break
		}
	}

	return string(buf)
}
