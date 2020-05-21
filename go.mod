module k8s-dt-node-labeller

go 1.13

replace github.com/platinasystems/fdt => github.com/adaptant-labs/fdt v1.0.2-0.20200521120921-a1cf022f5f28

require (
	github.com/go-logr/logr v0.1.0
	github.com/platinasystems/fdt v1.0.1
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	sigs.k8s.io/controller-runtime v0.5.2
)
