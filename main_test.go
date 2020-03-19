package main

import "testing"

func TestPropertyNormalize(t *testing.T) {
	p := "xlnx,zynq-7000"
	e := "xilinx-zynq-7000"
	n := propertyNormalize(p)

	if n != "xilinx-zynq-7000" {
		t.Errorf("Property value incorrect, got \"%s\", expected: \"%s\"", n, e)
	}
}
