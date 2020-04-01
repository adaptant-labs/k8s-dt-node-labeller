package main

import (
	"fmt"
	"strings"
)

var (
	labelNamespace = "devicetree.org"
)

func createLabelPrefix(name string, experimental bool) string {
	var s string

	if experimental {
		s = "beta."
	} else {
		s = ""
	}

	return fmt.Sprintf("%s%s/%s", s, labelNamespace, name)
}

func matchesLabelPrefix(s string) bool {
	return strings.HasPrefix(s, "beta." + labelNamespace) || strings.HasPrefix(s, labelNamespace)
}