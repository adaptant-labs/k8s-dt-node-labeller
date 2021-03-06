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

type CompatInfo map[string]int

var (
	compatMap              = make(CompatInfo)
	log                    = logf.Log.WithName("k8s-dt-node-labeller")
	defaultNodes           = []string{"cpu", "gpu" }
	featureFilesDir        = "/etc/kubernetes/node-feature-discovery/features.d/"
	nfdFeaturesFile        = "devicetree-features"
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
		// Match nodes with an exact name match
		t.MatchNode(node, walkNode)

		// Detect node names matching either of the following formats:
		//	name@<address>
		//	name<id>@<address>
		regex := fmt.Sprintf("^(%s)([0-9]+|)@.*$", node)
		t.EachNodeMatching(regex, walkNode)
	}

	return nil
}

func (ci CompatInfo) writeNfdFeatures() error {
	// Create the directory if it doesn't exist
	err := os.MkdirAll(featureFilesDir, os.ModePerm)
	if err != nil {
		return err
	}

	// Create the features file
	featuresFilePath := featureFilesDir + nfdFeaturesFile
	features, err := os.Create(featuresFilePath)
	if err != nil {
		return err
	}
	defer features.Close()

	fmt.Printf("Writing out discovered features to %s\n", featuresFilePath)

	// Write out each feature
	for k, v := range compatMap {
		label := fmt.Sprintf("%s=%d\n", createLabelPrefix(k, true), v)
		_, err := features.WriteString(label)
		if err != nil {
			return err
		}
	}

	features.Sync()
	return nil
}

func (ci CompatInfo) dumpFeatures() {
	fmt.Printf("Discovered the following devicetree properties:\n\n")

	// Iterate over the parsed map
	for k, v := range compatMap {
		fmt.Printf("%s: %d\n", createLabelPrefix(k, true), v)
	}
}

func appendNodeIfNotExist(nodes []string, name string) []string {
	for _, node := range nodes {
		if node == name {
			return nodes
		}
	}

	return append(nodes, name)
}

func appendNodesIfNotExist(nodes[] string, names ...string) []string {
	for _, name := range names {
		nodes = appendNodeIfNotExist(nodes, name)
	}

	return nodes
}

func getNodeName() (string, error) {
	// Within the Kubernetes Pod, the hostname provides the Pod name, rather than the node name, so we pass in the
	// node name via the NODE_NAME environment variable instead.
	nodeName := os.Getenv("NODE_NAME")
	if len(nodeName) > 0 {
		return nodeName, nil
	}

	// If the NODE_NAME environment variable is unset, fall back on hostname matching (e.g. when running outside of
	// a Kubernetes deployment).
	return os.Hostname()
}

func main() {
	var n string
	var d bool
	var nfd bool

	flag.BoolVar(&d, "d", false, "Display detected devicetree nodes")
	flag.BoolVar(&nfd, "f", false, "Write detected features to NFD features file")
	flag.StringVar(&n, "n", "", "Additional devicetree node names")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "devicetree Node Labeller for Kubernetes\n")
		fmt.Fprintf(os.Stderr, "Usage: k8s-dt-node-labeller [flags] [-n devicetree nodes...]\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	tail := flag.Args()

	// By default check for top-of-tree compatible strings
	nodeNames := []string{"/"}

	// And any specified default nodes
	nodeNames = append(nodeNames, defaultNodes...)

	// Include any additional node names specified
	if len(n) > 0 {
		nodeNames = appendNodeIfNotExist(nodeNames, n)

		if len(tail) > 0 {
			nodeNames = appendNodesIfNotExist(nodeNames, tail...)
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

	if d {
		compatMap.dumpFeatures()
		os.Exit(0)
	}

	if nfd {
		err = compatMap.writeNfdFeatures()
		if err != nil {
			fmt.Printf("Failed to write feature labels: %s\n", err.Error())
			os.Exit(1)
		}

		os.Exit(0)
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
	// tagged as part of the node's metadata.
	pred := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			hostname, err := getNodeName()
			if err != nil {
				return false
			}
			return hostname == e.Meta.GetName()
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
