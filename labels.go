package main

import (
	"fmt"
	"strings"
)

var (
	groupLabelAlias = map[string]map[string][]string {
		"accelerator":	{
			"gpu": { "nvidia-tegra210", },
			"fpga": { "xilinx-zynq7000", },
		},
	}
)

func createLabelPrefix(name string, experimental bool, dt bool) string {
	var s string

	if experimental {
		s = "beta."
	} else {
		s = ""
	}

	if dt {
		return fmt.Sprintf("%sadaptant.io/dt.%s", s, name)
	}

	return fmt.Sprintf("%sadaptant.io/%s", s, name)
}

func makeGroupedLabel(label string, members ...string) string {
	return fmt.Sprintf("%s=%s", label, strings.Join(members, ","))
}

// Generate labels in a comma delimited form of group=type, type=model.
// e.g. accelerator=gpu,fpga gpu=nvidia-tegra210 fpga=xilinx-zynq7000
// This enables labels to be selected either on the class of accelerator,
// or a specific model.
func generateLabels() []string {
	var labels []string
	for groupType, groupMap := range groupLabelAlias {
		var groups []string

		for memberType, possibleMembers := range groupMap {
			var members []string

			matched := false
			for k, _ := range compatMap {
				for i := 0; i < len(possibleMembers); i++ {
					if k == possibleMembers[i] {
						matched = true
						members = append(members, possibleMembers[i])
						break
					}
				}

				if matched == true {
					break
				}
			}

			if matched == true {
				groups = append(groups, memberType)
				labels = append(labels, makeGroupedLabel(memberType, members...))
			}
		}

		if len(labels) > 0 {
			labels = append(labels, makeGroupedLabel(groupType, groups...))
		}
	}

	return labels
}