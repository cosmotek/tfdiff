package terraform

import (
	"os/exec"
	"strings"
)

func GetWorkspace() (string, error) {
	_, err := WhichTerraform()
	if err != nil {
		return "", err
	}

	output, err := exec.Command("terraform", "workspace", "show").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(output), "\n"), nil
}
