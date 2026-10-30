package terraform

import (
	"os/exec"
	"strings"
)

func GetWorkspace() (string, error) {
	bin, err := WhichTerraform()
	if err != nil {
		return "", err
	}

	output, err := exec.Command(bin, "workspace", "show").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(output), "\n"), nil
}
