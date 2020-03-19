package main

import (
	"fmt"
)

var (
	labelNamespace = "sodalite.eu"
)

func createLabelPrefix(name string, experimental bool) string {
	var s string

	if experimental {
		s = "beta."
	} else {
		s = ""
	}

	return fmt.Sprintf("%s%s/dt.%s", s, labelNamespace, name)
}
