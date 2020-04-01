# Device Tree Kubernetes Node Labeller

This tool provides a custom Kubernetes controller for automatically labelling nodes with [devicetree] properties.

It is inspired by and re-uses much of the Kubernetes controller plumbing from the [amdgpu-node-labeller].

## Use Cases

`k8s-dt-node-labeller` was developed in order to facilitate targeted deployments into hybrid (e.g. armhf, arm64)
Kubernetes clusters, as well as for targeting heterogeneous accelerators in Edge deployments.

## Usage

The node labeller is further expected to be run node-local, and will need to be invoked on each individual node
requiring its own specific devicetree parsing and labelling.

By default, `compatible` strings from the top-level `/` node are discovered and converted to node labels.

```
$ k8s-dt-node-labeller
Discovered the following devicetree properties:

beta.devicetree.org/nvidia-jetson-nano: 1
beta.devicetree.org/nvidia-tegra210: 1
```

additional node specifications are possible via the `-n` flag, as below:

```
$ k8s-dt-node-labeller -n gpu pwm-fan
Discovered the following devicetree properties:

beta.devicetree.org/nvidia-jetson-nano: 1
beta.devicetree.org/nvidia-tegra210: 1
beta.devicetree.org/nvidia-tegra210-gm20b: 1
beta.devicetree.org/nvidia-gm20b: 1
beta.devicetree.org/pwm-fan: 1
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