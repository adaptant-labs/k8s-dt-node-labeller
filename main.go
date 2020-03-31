package main

import (
	"flag"
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

// Normalize 'manufacturer,model' property names to 'manufacturer-model'
func propertyNormalize(property string) string {
	s := strings.Split(property, ",")

	// If there is no matching separator, return the property as-is
	if len(s) == 1 {
		return property
	}

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

func parseDeviceTree(nodeNames []string) error {
	if _, err := os.Stat("/proc/device-tree"); err != nil {
		return fmt.Errorf("no valid device tree configuration found")
	}

	t := fdt.DefaultTree()
	if t == nil {
		panic("Failed to parse device tree")
	}

	for _, node := range nodeNames {
		t.MatchNode(node, walkNode)
	}

	return nil
}

func main() {
	var n string

	flag.StringVar(&n, "n", "", "Additional devicetree node names")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "devicetree Node Labeller for Kubernetes\n")
		fmt.Fprintf(os.Stderr, "Usage: k8s-dt-node-labeller [flags]\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	tail := flag.Args()

	// By default check for top-of-tree compatible strings
	nodeNames := []string{ "/" }

	// Include any additional node names specified
	if len(n) > 0 {
		nodeNames = append(nodeNames, n)

		if len(tail) > 0 {
			nodeNames = append(nodeNames, tail...)
		}
	} else {
		// Check for any invalid or unhandled args
		if len(tail) > 0 {
			flag.Usage()
			os.Exit(1)
		}
	}

	err := parseDeviceTree(nodeNames)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Discovered the following devicetree properties:\n\n")

	// Iterate over the parsed map
	for k, v := range compatMap {
		fmt.Printf("%s: %d\n", createLabelPrefix(k, true), v)
	}
}
