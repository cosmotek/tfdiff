package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// binaryCandidates lists the IaC CLIs tfdiff can drive, in priority order.
// OpenTofu is preferred, with HashiCorp Terraform as a fallback.
var binaryCandidates = []string{"tofu", "terraform"}

// WhichTerraform returns the path to the first available IaC CLI (tofu or
// terraform), or an error if neither is on $PATH.
func WhichTerraform() (string, error) {
	for _, name := range binaryCandidates {
		path, err := exec.LookPath(name)
		if err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("neither tofu nor terraform was found on $PATH")
}

func PullState() (PullStateOutput, error) {
	bin, err := WhichTerraform()
	if err != nil {
		return PullStateOutput{}, err
	}

	data, err := exec.Command(bin, "state", "pull").Output()
	if err != nil {
		return PullStateOutput{}, fmt.Errorf("failed to execute '%s state pull': %w", bin, err)
	}

	output := PullStateOutput{}
	err = json.Unmarshal(data, &output)
	if err != nil {
		return PullStateOutput{}, fmt.Errorf("failed to parse terraform state json: %w", err)
	}

	return output, nil
}

// LoadState parses a state file previously captured with `tofu state pull` (or
// `terraform state pull`), for diffing several projects' states at once without
// re-running the CLI against each one.
func LoadState(path string) (PullStateOutput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PullStateOutput{}, fmt.Errorf("failed to read state file %s: %w", path, err)
	}

	output := PullStateOutput{}
	err = json.Unmarshal(data, &output)
	if err != nil {
		return PullStateOutput{}, fmt.Errorf("failed to parse state json %s: %w", path, err)
	}

	return output, nil
}

type PullStateOutput struct {
	Version          int    `json:"version"`
	TerraformVersion string `json:"terraform_version"`
	Serial           int    `json:"serial"`
	Lineage          string `json:"lineage"`
	Outputs          struct {
	} `json:"outputs"`
	Resources []struct {
		Module    string `json:"module"`
		Mode      string `json:"mode"`
		Type      string `json:"type"`
		Name      string `json:"name"`
		Provider  string `json:"provider"`
		Instances []struct {
			SchemaVersion int `json:"schema_version"`
			// Only the resource identifier is read downstream (Arn for AWS in
			// cmd/aws.go, ID for Azure in cmd/azure.go). Every other attribute
			// varies by provider and resource type, so pinning them to concrete
			// Go types only invites unmarshal failures -- e.g. maintenance_window
			// is a string on AWS RDS but a block (JSON array) on several Azure
			// resources. encoding/json ignores keys we omit, which is what we want.
			Attributes struct {
				Arn string `json:"arn"`
				ID  string `json:"id"`
			} `json:"attributes"`
			Private      string   `json:"private"`
			Dependencies []string `json:"dependencies"`
		} `json:"instances"`
	} `json:"resources"`
	CheckResults interface{} `json:"check_results"`
}
