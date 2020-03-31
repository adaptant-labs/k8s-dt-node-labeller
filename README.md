# Device Tree Kubernetes Node Labeller

This tool automatically labels nodes with [devicetree] properties, useful for targeted deployments into hybrid
Kubernetes clusters, as well as for targeting heterogeneous accelerator resources in Edge deployments.

[devicetree]: https://www.devicetree.org

By default, `compatible` strings from the top-level `/` node are discovered and converted to node labels.

```
$ k8s-dt-node-labeller
Discovered the following devicetree properties:

beta.devicetree.org/nvidia-jetson-nano: 1
beta.devicetree.org/nvidia-tegra210: 1
```

node specifications are possible via the `-n` flag, as below:

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

[SODALITE]: https://www.sodalite.eu

## License

`k8s-dt-node-labeller` is licensed under the terms of the Apache 2.0 license, the full
version of which can be found in the LICENSE file included in the distribution.
