package main

import (
	"fmt"
	"testing"
)

func TestLabelPrefix(t *testing.T) {
	devName := "xilinx-zynq-7000"
	labelPrefix := "beta.devicetree.org"
	expected := fmt.Sprintf("%s/%s", labelPrefix, devName)
	s := createLabelPrefix(devName, true)
	if s != expected {
		t.Errorf("Label incorrect, got \"%s\", expected: \"%s\"", s, expected)
	}
}
