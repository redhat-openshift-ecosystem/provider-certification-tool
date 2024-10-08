# User Guide

Welcome to the user documentation for the OpenShift Provider Compatibility Tool (OPCT)!

The OPCT is used to validate an OpenShift/OKD installation on an infrastructure or hardware provider is in conformance with required e2e suites.

> Note: This document is under `preview` release and it's in constant improvement.

Table Of Contents:

- [Process Overview](#process)
- [Prerequisites](#prerequisites)
    - [Standard Environment](#standard-env)
        - [Setup Dedicated Node](#standard-env-setup-node)
        - [Setup MachineConfigPool (upgrade mode)](#standard-env-setup-mcp)
        - [Testing in a Disconnected Environment](#disconnected-env-setup)
    - [Privilege Requirements](#priv-requirements)
- [Install](#install)
    - [Prebuilt Binary](#install-bin)
    - [Build from Source](#install-source)
- [Usage](#usage)
    - [Run tool](#usage-run)
        - [Default Run mode](#usage-run-regular)
        - [Run 'upgrade' mode](#usage-run-upgrade)
        - [Optional parameters](#usage-run-optional)
    - [Check status](#usage-check)
    - [Collect the results](#usage-retrieve)
    - [Check the Results](#usage-results)
    - [Review the Report](#usage-results)
    - [Submit the Results](#submit-results)
    - [Environment Cleanup](#usage-destroy)
- [Troubleshooting](#troubleshooting)
- [Feedback](#feedback)

## Process Overview <a name="process"></a>

This section describes the steps of the process when submiting the results to Red Hat Partner.
If the goal is not sharing the results to Red Hat, you can go to the next section.

Overview of the process to apply the results to the Red Hat Partner support case:

0. Install an OpenShift cluster on **the version desired** to be validated
1. Prepare the OpenShift cluster to run the validated environment
2. Download and install the OPCT
3. Run the OPCT
4. Monitor tests 
5. Gather results
6. Share the results archive with Red Hat Partner support case

It's not uncommon for specific tests to occasionally fail.  As a result, you may be asked by Support Engineers to repeat the process a few times depending on the results.

Finally, you will be asked to manually upgrade the cluster to the next minor release.

More detail on each step can be found in the sections further below.

## Prerequisites <a name="prerequisites"></a>

A Red Hat OpenShift 4 cluster must be [installed](https://docs.openshift.com/container-platform/latest/installing/index.html) before validation can begin. The OpenShift cluster must be installed on your infrastructure as if it were a production environment. Ensure that each feature of your infrastructure you plan to support with OpenShift is configured in the cluster (e.g. Load Balancers, Storage, special hardware).

OpenShift supports the following topologies:

- Highly available OpenShift Container Platform cluster (**HA**): Three control plane nodes with any number of compute nodes.
- A three-node OpenShift Container Platform cluster (**Compact**): A compact cluster that has three control plane nodes that are also compute nodes.
- A single-node OpenShift Container Platform cluster (**SNO**): A node that is both a control plane and compute.

OPCT is tested in the following topologies - uncovered topologies(flagged as TBD) is not supported by the tool in the validation process:

| OCP Topology/ARCH | OPCT Initial version | OPCT Execution mode |
| -- | -- | -- |
| HA/amd64 | v0.1 | regular(v0.1+), upgrade(v0.3+), disconnect(v0.4+) |
| HA/arm64 | v0.5 | all |
| Compact/amd64 | TBD([OPCT-193](https://issues.redhat.com/browse/OPCT-193)) | -- |
| SNO/amd64 | TBD([OPCT-30](https://issues.redhat.com/browse/OPCT-30)) | -- |

!!! info "Unsupported Topologies"
    You must be able to run the tool in unsupported topologies when the required configuration is set,
    although the report provided by the tool may not be calibrated or expected results to cover the
    formal validation process when applying to Red Hat OpenShift programs for Partners.

OpenShift Platform Type supported by OPCT:

| Platform Type | OCP Supported versions |
| -- | -- |
| None | v0.1+ |
| External | v0.5+(preview) |

!!! info "Unsupported Platform Type"
    You must be able to run the tool in other platform types when the required configuration is set,
    although the reports may not be calibrated or with expected results to cover the
    full execution, leading to failures of platform-specific e2e tests that requires special configuration
    or credentials.

The matrix below describes the OpenShift and OPCT versions supported for each release and features:

| OPCT [version](releases) | OCP tested versions | OPCT Execution mode |
| -- | -- | -- |
| v0.5.x | 4.14-4.17 | regular, upgrade, disconnected |
| v0.4.x | 4.10-4.13 | regular, upgrade, disconnected |
| v0.3.x | 4.9-4.12 | regular, upgrade |
| v0.2.x | 4.9-4.11 | regular |
| v0.1.x | 4.9-4.11 | regular |

It's highly recommended to use the latest OPCT version.

[releases]:https://github.com/redhat-openshift-ecosystem/opct/releases

### Standard Environment <a name="standard-env"></a>

A dedicated compute node should be used to avoid disruption of the test scheduler. Otherwise, the concurrency between resources scheduled on the cluster, e2e-test manager (aka openshift-tests-plugin), and other stacks like monitoring can disrupt the test environment, leading to unexpected results, like the eviction of plugins or aggregator server (`sonobuoy` pod).

The dedicated node environment cluster size can be adjusted to match the table below. Note the differences in the `Dedicated Test` machine:

| Machine       | Count | CPU | RAM (GB) | Storage (GB) |
| ------------- | ----- | --- | -------- | ------------ |
| Bootstrap     | 1     | 4   | 16       | 100          |
| Control Plane | 3     | 4   | 16       | 100          |
| Compute       | 3     | 4   | 16       | 100          |
| Dedicated Test| 1     | 4   | 8        | 100          |

*Note: These requirements are higher than the [minimum requirements](https://docs.openshift.com/container-platform/latest/installing/installing_bare_metal/installing-bare-metal.html#installation-minimum-resource-requirements_installing-bare-metal) for a regular cluster (non-validation) in OpenShift product documentation due to the resource demand of the conformance environment.*

#### Environment Setup: Dedicated Node <a name="standard-env-setup-node"></a>

The `Dedicated Node` is a regular worker with additional label and taints to run the OPCT environment.

The following requirements must be satisfied:

1. Choose one node with at least 8GiB of RAM and 4 CPU
2. Label node with `node-role.kubernetes.io/tests=""` (conformance-related pods will schedule to dedicated node)
3. Taint node with `node-role.kubernetes.io/tests="":NoSchedule` (prevent other pods from running on dedicated node)

Example setting up the dedicated node:

```shell
oc label node <node_name> node-role.kubernetes.io/tests=""
oc adm taint node <node_name> node-role.kubernetes.io/tests="":NoSchedule
```

#### Setup MachineConfigPool for upgrade tests <a name="standard-env-setup-mcp"></a>

**Note**: The `MachineConfigPool` should be created only when the OPCT execution mode (`--mode`) is `upgrade`. If you are not running upgrade tests, please skip this section.

One `MachineConfigPool`(MCP) with the name `opct` must be created, selecting the dedicated node labels. The MCP must be `paused`, thus the node running the validation environment will not be restarted while the cluster is upgrading, avoiding disruptions to the conformance results.

You can create the `MachineConfigPool` by running the following command:

```bash
cat << EOF | oc create -f -
apiVersion: machineconfiguration.openshift.io/v1
kind: MachineConfigPool
metadata:
  name: opct
spec:
  machineConfigSelector:
    matchExpressions:
    - key: machineconfiguration.openshift.io/role,
      operator: In,
      values: [worker,opct]
  nodeSelector:
    matchLabels:
      node-role.kubernetes.io/tests: ""
  paused: true
EOF
```

Make sure the `MachineConfigPool` has been created correctly:

```bash
oc get machineconfigpool opct
```

#### Testing in a Disconnected Environment <a name="disconnected-env-setup"></a>

The OPCT requires numerous images during the setup and execution of tests.
See [User Installation Guide - Disconnected Installations](./user-installation-disconnected.md) for details
on how to configure a mirror registry and how to run the OPCT to rely on the mirror
registry for images.

### Privilege Requirements <a name="priv-requirements"></a>

A user with [cluster administrator privilege](https://docs.openshift.com/container-platform/latest/authentication/using-rbac.html#creating-cluster-admin_using-rbac) must be used to run the tool. You also use the default `kubeadmin` user if you wish.

## Install <a name="install"></a>

The OPCT is shipped as a single executable binary which can be downloaded from [the Project Releases page](https://github.com/redhat-openshift-ecosystem/opct/releases). Choose the latest version and the architecture of the node (client) you will execute the tool, then download the binary.

The tool can be used from any system with access to API to the OpenShift cluster under test.

```sh
VERSION=v0.4.1
BINARY=opct-linux-amd64
wget -O opct "https://github.com/redhat-openshift-ecosystem/opct/releases/download/${VERSION}/${BINARY}"
chmod u+x ./opct
```

!!! warning "OPCT binary"
    The tool has been renamed from `openshift-provider-cert` to `opct` starting at v0.5.x release.

    The documentation is adapted renaming the references to the older name.

## Usage <a name="usage"></a>

### Run conformance tests <a name="usage-run"></a>

Requirements:

- You have set the dedicated node
- You have installed OPCT

#### Run the default execution mode <a name="usage-run-regular"></a>

Create and run the validation environment (detaching the terminal/background):

```sh
./opct run
```

Optionally you can watch the execution using `--watch`:

```sh
./opct run --watch
```

#### Run the `upgrade` mode <a name="usage-run-upgrade"></a>

The `upgrade` mode runs the OpenShift cluster updates to the `4.y+1` version, then the regular conformance suites will be executed (Kubernetes and OpenShift). This mode was created to validate the entire update process, and to make sure the target OCP release is validated on the conformance suites.

> Note: If you will submit the results to Red Hat Partner Support, you must have Validated the installation on the initial release using the regular execution. For example, to submit the upgrade tests for 4.11->4.12, you must have submitted the regular tests for 4.11. If you have any questions, ask your Red Hat Partner Manager.

Requirements for running the `upgrade` mode:

- You have created the `MachineConfigPool` with name `opct`
- You have installed the OpenShift client locally (`oc`) - or have noted the Image `Digest` of the target release
- You must choose the next release of Y-stream (`4.Y+1`) supported by your current release. (See [update graph](https://access.redhat.com/labs/ocpupgradegraph/update_path))

```sh
./opct run --mode=upgrade --upgrade-to-image=$(oc adm release info 4.Y+1.Z -o jsonpath={.image})
```

#### Run with the Disconnected Mirror registry<a name="usage-run-disconnected"></a>

Tests are able to be run in a disconnected environment through the use of a mirror registry.

Requirements for running tests with a disconnected mirror registry:

- Disconnected Mirror Image Registry created
- [Private cluster Installed](https://docs.openshift.com/container-platform/latest/installing/installing_bare_metal/installing-restricted-networks-bare-metal.html)
- [You created a registry on your mirror host](https://docs.openshift.com/container-platform/latest/installing/disconnected_install/installing-mirroring-installation-images.html#installing-mirroring-installation-images)


To run tests such that they use images hosted by the Disconnected Mirror registry:

```sh
./opct run --image-repository ${TARGET_REPO}
```

### Check status <a name="usage-check"></a>

```sh
./opct status

# OR Keep watch open until completion

./opct status -w
```

### Collect the results <a name="usage-retrieve"></a>

The results must be retrieved from the OpenShift cluster under test using:

```sh
./opct retrieve

# OR save to the target directory

./opct retrieve ./destination-dir/
```

The file must be saved locally.

### Check the results archive <a name="usage-results"></a>

You can see a summarized view of the results using:

```sh
./opct results <retrieved-archive>.tar.gz
```

#### Review the report <a name="usage-report"></a>

Review the data generated by report:

```sh
./opct report <retrieved-archive>.tar.gz
```

### Submit the results archive <a name="submit-results"></a>

How to submit OPCT results from the validated environment:

- Log in to the [Red Hat Connect Portal](https://connect.redhat.com/login).
- Go to [`Support > My support tickets > Create Case`](https://connect.redhat.com/support/technology-partner/#/case/new).
- In the `Request Category` step, select `Product Certification`.
- In the `Product Selection` step, for the Product field, select `OpenShift Container Platform` and select the Version you are using.
- Click `Next` to continue.
- In the `Request Details` step, in the `Request Summary` field, specify `[VCSP] OPCT Test Results <provider name>` and provide any additional details in the `Please add description` field.
- Click `Next` to continue.
- Click `Submit` when you have completed all the required information.
- A Product Certification ticket will be created, and please follow the instructions provided to add the test results and any other related material for us to review.
- Go to [`Support > My support tickets`](https://connect.redhat.com/support/technology-partner/#/case/list) to find the case and review status and/or to add comments to the case.

Required files to attach to a NEW support case:

- Attach the detailed Deployment Document describing how the cluster is installed, architecture, flavors and additional/specific configurations from your validated Cloud Provider.
- Download, review and attach the [`user-installation-checklist.md`](./user-installation-checklist.md) to the case.
- Attach the `<retrieved-archive>.tar.gz` result file to the case.


### Environment Cleanup <a name="usage-destroy"></a>

Once the validation process is complete and you are comfortable with destroying the test environment:

```sh
./opct destroy
```

You will need to destroy the OpenShift cluster under test separately. 

## Troubleshooting Helper

Check also the documents below that might help while investigating the results and failures of the validation process:

- [Troubleshooting Guide](./troubleshooting-guide.md)
- [Installation Review](./user-installation-review.md)

## Feedback <a name="feedback"></a>

If you are a community user and found any bugs or issues, you can open a [new GitHub issue](https://github.com/redhat-openshift-ecosystem/opct/issues/new).

If you are under validation process and are looking for guidance or feedback, please reach out to your Red Hat Partner Manager to assist you with the conformance process.
