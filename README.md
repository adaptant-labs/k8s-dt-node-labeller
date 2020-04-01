# Device Tree Kubernetes Node Labeller

This tool provides a custom Kubernetes controller for automatically labelling nodes with [devicetree] properties.

It is inspired by and re-uses much of the Kubernetes controller plumbing from the [amdgpu-node-labeller].

## Use Cases

`k8s-dt-node-labeller` was developed in order to facilitate targeted deployments into hybrid (e.g. armhf, arm64)
Kubernetes clusters, as well as for targeting heterogeneous accelerators in Edge deployments.

## Usage

The node labeller is expected to be run node-local, and will need to be invoked on each individual node
requiring its own specific devicetree parsing and labelling.

```
$ k8s-dt-node-labeller --help
devicetree Node Labeller for Kubernetes
Usage: k8s-dt-node-labeller [flags] [-n devicetree nodes...]

  -d	Display detected devicetree nodes
  -kubeconfig string
    	Paths to a kubeconfig. Only required if out-of-cluster.
  -n string
    	Additional devicetree node names

```

By default, `compatible` strings from the top-level `/` node are discovered and converted to node labels.

```
$ k8s-dt-node-labeller -d
Discovered the following devicetree properties:

beta.devicetree.org/nvidia-jetson-nano: 1
beta.devicetree.org/nvidia-tegra210: 1
```

additional node specifications are possible via the `-n` flag, as below:

```
$ k8s-dt-node-labeller -d -n gpu pwm-fan
Discovered the following devicetree properties:

beta.devicetree.org/nvidia-jetson-nano: 1
beta.devicetree.org/nvidia-tegra210: 1
beta.devicetree.org/nvidia-tegra210-gm20b: 1
beta.devicetree.org/nvidia-gm20b: 1
beta.devicetree.org/pwm-fan: 1
```

In order to submit the labels to the cluster, ensure a valid kubeconfig can be found (e.g. by setting the `KUBECONFIG`
environment variable, or through specifying the location with the `-kubeconfig` parameter) and execute the node
labeller as-is:

```
$ k8s-dt-node-labeller 
{"level":"info","ts":1585701002.694782,"logger":"k8s-dt-node-labeller.entrypoint","msg":"setting up manager"}
{"level":"info","ts":1585701003.1179338,"logger":"controller-runtime.metrics","msg":"metrics server is starting to listen","addr":":8080"}
{"level":"info","ts":1585701003.118381,"logger":"k8s-dt-node-labeller.entrypoint","msg":"Setting up controller"}
{"level":"info","ts":1585701003.1185818,"logger":"k8s-dt-node-labeller.entrypoint","msg":"starting manager"}
{"level":"info","ts":1585701003.1190712,"logger":"controller-runtime.manager","msg":"starting metrics server","path":"/metrics"}
{"level":"info","ts":1585701003.1193638,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"k8s-dt-node-labeller","source":"kind source: /, Kind="}
{"level":"info","ts":1585701003.2218263,"logger":"controller-runtime.controller","msg":"Starting Controller","controller":"k8s-dt-node-labeller"}
{"level":"info","ts":1585701003.222448,"logger":"controller-runtime.controller","msg":"Starting workers","controller":"k8s-dt-node-labeller","worker count":1}
...
```

After which the node labels can be viewed from `kubectl`:

```
$ kubectl describe node jetson-nano
Name:               jetson-nano
Roles:              <none>
Labels:             beta.devicetree.org/nvidia-jetson-nano=1
                    beta.devicetree.org/nvidia-tegra210=1
                    beta.kubernetes.io/arch=arm64
                    beta.kubernetes.io/instance-type=k3s
                    beta.kubernetes.io/os=linux
                    k3s.io/hostname=jetson-nano
                    k3s.io/internal-ip=192.168.xxx.xxx
                    kubernetes.io/arch=arm64
                    kubernetes.io/hostname=jetson-nano
                    kubernetes.io/os=linux
                    node.kubernetes.io/instance-type=k3s
                    ...
```

## Acknowledgements

This project has received funding from the European Union’s Horizon 2020 research and innovation programme under grant
agreement No 825480 ([SODALITE]).

## License

`k8s-dt-node-labeller` is licensed under the terms of the Apache 2.0 license, the full
version of which can be found in the LICENSE file included in the distribution.

[devicetree]: https://www.devicetree.org
[SODALITE]: https://www.sodalite.eu
[amdgpu-node-labeller]: https://github.com/RadeonOpenCompute/k8s-device-plugin/tree/master/cmd/k8s-node-labeller