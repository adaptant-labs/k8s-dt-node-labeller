package main

import (
	"fmt"
	"github.com/platinasystems/fdt"
	"os"
	"strings"
)

var (
	compatMap = make(map[string]int)
	vendorNormalizationMap = map[string]string {
		"xlnx": "xilinx",
	}
)

func vendorNormalize(vendorName string) string {
	if normalized, ok := vendorNormalizationMap[vendorName]; ok {
		return normalized
	}

	return vendorName
}

func propertyNormalize(property string) string {
	s := strings.Split(property, ",")
	return fmt.Sprintf("%s-%s", vendorNormalize(s[0]), s[1])
}

func getCompatStrings(n *fdt.Node) []string {
	s := strings.Split(string(n.Properties["compatible"]), "\x00")
	s = s[:len(s)-1]
	return s
}

func walkNode(n *fdt.Node) {
	compatStrings := getCompatStrings(n)
	for _, v := range compatStrings {
		compatMap[propertyNormalize(v)] += 1
	}
}

func main() {
	if _, err := os.Stat("/proc/device-tree"); err != nil {
		panic("No valid device tree configuration found")
	}
	t := fdt.DefaultTree()
	if t == nil {
		panic("Failed to parse device tree")
	}

	t.MatchNode("/", walkNode)

	for i := 1; i < len(os.Args); i++ {
		t.MatchNode(os.Args[i], walkNode)
	}

	for k, v := range compatMap {
		fmt.Printf("%s: %d\n", createLabelPrefix(k, true), v)
	}
}
