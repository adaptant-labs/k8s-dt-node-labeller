package main

import (
	"fmt"
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
