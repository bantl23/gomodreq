package modinfo

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// GetModInfo returns module information from go list cli command
func GetModInfo() ([]*ModulePublic, error) {
	list, err := exec.Command("go", "list", "-m", "-f", "{{.Path}}", "all").Output()
	if err != nil {
		return nil, err
	}

	miSlice := make([]*ModulePublic, 0)
	scanner := bufio.NewScanner(strings.NewReader(string(list)))
	for scanner.Scan() {
		result, err := exec.Command("go", "list", "-m", "-u", "-json", scanner.Text()).Output()
		if err != nil {
			return nil, fmt.Errorf("error getting mod info %s [%+v]", scanner.Text(), err)
		}
		mi := ModulePublic{}
		err = json.Unmarshal(result, &mi)
		if err != nil {
			return nil, fmt.Errorf("error parsing json [%+v]", err)
		}
		miSlice = append(miSlice, &mi)

	}
	return miSlice, nil
}
