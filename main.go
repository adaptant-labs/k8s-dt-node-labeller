package main

import (
	"flag"
	"fmt"
	"github.com/platinasystems/fdt"
	corev1 "k8s.io/api/core/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strconv"
	"strings"
)

var (
	compatMap              = make(map[string]int)
	log                    = logf.Log.WithName("k8s-dt-node-labeller")
	vendorNormalizationMap = map[string]string{
		"xlnx": "xilinx",
	}
)

// Convert the compatMap to the label format expected by the K8s Reconciler API
func generateLabels(experimental bool) map[string]string {
	m := make(map[string]string)
	for k, v := range compatMap {
		m[createLabelPrefix(k, experimental)] = strconv.Itoa(v)
	}

	return m
}

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
		return fmt.Errorf("failed to parse device tree")
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
	nodeNames := []string{"/"}

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

	logf.SetLogger(zap.New(zap.UseDevMode(false)))
	entryLog := log.WithName("entrypoint")

	// Setup a Manager
	entryLog.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		entryLog.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	// Setup a new controller to Reconcile Node labels
	entryLog.Info("Setting up controller")
	c, err := controller.New("k8s-dt-node-labeller", mgr, controller.Options{
		Reconciler: &reconcileNodeLabels{client: mgr.GetClient(),
			log:    log.WithName("reconciler"),
			labels: generateLabels(true),
		},
	})
	if err != nil {
		entryLog.Error(err, "unable to set up individual controller")
		os.Exit(1)
	}

	// By default, only run node-local. This is achieved by matching the
	// hostname of the system we are running on against the hostname
	//tagged as part of the node's metadata.
	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			hostname, err := os.Hostname()
			if err != nil {
				return false
			}
			if hostname == e.Meta.GetName() {
				entryLog.Info("Labelling", hostname)
				return true
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	// Watch Nodes and enqueue Nodes object key
	if err := c.Watch(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{}, &pred); err != nil {
		entryLog.Error(err, "unable to watch Node")
		os.Exit(1)
	}

	entryLog.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		entryLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}
