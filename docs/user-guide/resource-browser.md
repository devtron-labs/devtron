# Resource Browser

## Introduction

The Devtron Resource Browser provides you a central interface to view and manage all your [Kubernetes objects](../../reference/glossary.md#objects) across clusters.  It helps you perform key actions like viewing logs, editing live manifests, and even creating/deleting resources directly from the user interface. This is especially useful for troubleshooting purposes as it supports multi-cluster too.

First, the Resource Browser shows you a list of clusters added to your Devtron setup. By default, it displays a cluster named '*default_cluster*' after the [initial setup](../../setup/install/README.md) is successful.

![Devtron Resource Browser - List of Clusters](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/resource-browser.jpg)

You can easily connect more clusters by clicking the **Add Cluster** button located at the top of the browser. This will take you to the [Cluster & Environments](../global-configurations/cluster-and-environments.md) configuration within [Global Configurations](../global-configurations/README.md).

You may click a cluster to view and manage all its resources as shown below.

![Resources within Cluster](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/resource-list.jpg)

---

## Overview

| Fields | Description |
| --- | --- |
| **Cluster** | User-defined name for the cluster in Devtron. E.g. `default-cluster` |
| **Connection Status** | Status of the cluster. The status can be either `Successful` or `Failed`. |
| **Nodes** | Shows the number of nodes in a cluster. |
| **Node Errors** | Shows the error in nodes. |
| **K8s Version** | Shows the version of Kubernetes cluster. |
| **CPU Capacity** | Shows the CPU capacity in your cluster in milicore. E.g., 8000m where 1000 milicore equals to 1 core. |
| **Memory Capacity** | Shows the memory capacity in your cluster in mebibytes. |

To see the details of resource allocation and usage of the cluster, click the particular cluster.

On the `Resource allocation and usage`, you can see the details of compute resources or resources.

* CPU resources
* Memory resources

If you specify a `request` and `limits` in the container resource manifest file, then the respective values will appear on the `Resource allocation and usage` section.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/clusters/resource-allocation-usage-latest.jpg)

---

## Browsing Resources 

---

## Nodes

You can see the list of nodes available in your cluster. Typically you have several nodes in a cluster; in a learning or resource-limited environment, you might have only one node.

The components on a typical node include the `kubelet`, a `container runtime`, and the `kube-proxy`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/clusters/cluster-nodes.jpg)

If you have multiple nodes, you can search a node by name or label in the search bar.

| Fields | Description |
| --- | --- |
| **Node** | User-defined name for the node in Devtron. E.g. `demo-new`.<br>Note: Two nodes cannot have the same name at the same time.</br> |
| **Status** | Status of a node. It can be either `Ready` or `Not Ready`. |
| **Roles** | Shows the roles of a node. |
| **Errors** | Shows the error in nodes. |
| **K8s Version** | Shows the version of Kubernetes cluster. |
| **No. of Pods** | Shows the number of namespaces or pods in a node. |
| **CPU Usage** | Shows the CPU consumption in a node. |
| **Mem Usage** | Shows the memory consumption in a node |
| **Age** | Shows the time that the pod has been running since the last restart. |

To display a parameter of a node, use the `Columns` on the right side, select the parameter you want to display from the drop-down list and click `Apply`.

### Manage Nodes

Your applications run on pods, and pods run on Nodes. But sometimes, Kubernetes scheduler cannot deploy a pod on a Node for several reasons. As an example, node is not ready, node is not reachable, network is unavailable etc.

Using the Devtron UI, you will be able to:
* [Debugging a Node](#debug-a-node)
* [Cordoning a Node](#cordon-a-node)
* [Draining a Node](#drain-a-node)
* [Tainting a Node](#taint-a-node)
* [Editing a Node Config](#edit-a-node-config)
* [Deleting a Node](#delete-a-node)

To see the summary of a node, click the specific node.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/clusters/summary-of-a-node.jpg)

### Debugging a Node

You can debug a node via Terminal by selecting your namespace and image from the list that has all CLI utilities like kubectl, helm, netshoot etc. or can use a custom image, which is publicly available.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/clusters/node-terminal.png)

* Go to the `Clusters` section from the left navigation pane.
* Select your cluster.
* Search a node by name or label in the search bar.
* On the node, click the ellipsis button and then click `Terminal`.
* Debug a node by selecting the terminal shell `bash` or `sh`.

### Cordoning a Node

Cordoning a node means making the node unschedulable. After cordoning a node, new Pods cannot be scheduled on this node.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/clusters/node-cordon.jpg)

* On the node, click the ellipsis button and then click `Cordon`.
* A dialog box with message `Cordoning this node will mark this node as unschedulable. By cordoning a node, you can be sure that no new pods will be scheduled on this node` is displayed.
* The status of the node shows `SchedulingDisabled` with `Unschedulable` parameter set as `true`.

Similarly, you can uncordon a node by clicking `Uncordon`. After a node is uncordoned, new Pods can be scheduled on the node.

### Draining a Node

Before performing maintenance on a node, draining a node evicts all of your pods safely from a node. Safe evictions allow the pod’s containers to gracefully terminate and honour the `PodDisruptionBudgets` you have specified (if relevant).

After the node is drained, all Pods (including those managed by DaemonSets) in the node will be automatically drained to other nodes in the cluster, and the drained node will be set to cordoned status.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/clusters/drain-node.jpg)

* On the node, click the ellipsis button and then click `Drain`.
* A dialog box with message `Drain will cordon off the node and evict all pods of the node` is displayed.
* Click `Drain Node`.

You can also select from the following conditions before draining a node:

| Name | Usage |
| --- | --- |
| **Grace Period** | Period of time in seconds given to each pod to terminate gracefully. If negative, the default value specified in the pod will be used. |
| **Delete empty directory data** | Enabling this field will delete the pods using empty directory data when the node is drained. |
| **Disable eviction (use with caution** | Enabling this field will force drain to use delete, even if eviction is supported. This will bypass checking `PodDisruptionBudgets`.<br>Note: Make sure to use with caution.</br> |
| **Force drain** | Enabling this field will force drain a node even if there are pods that do not declare a controller. |
| **Ignore DaemonSets** | Enabling this field will ignore DaemonSet-managed pods. |

### Tainting a Node

Taints are `key:value` pairs associated with effect. After you add taints to nodes, you can set tolerations on a pod to allow the pod to be scheduled to nodes with certain taints. When you taint a node, it will repel all the pods except those that have a toleration for that taint. A node can have one or many taints associated with it.

**Note**: Make sure to check taint validations before you add a taint.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/clusters/edit-taints.jpg)

* On the node, click the ellipsis button and then click `Edit taints`.
* Click `Add taint`.
* On the `Key` and `Value` fields, enter the `key:value` pairs and select the [taint effect](#taint-effects) from the drop-down list.
* Click `Save`.
* You can also delete the added taint by clicking delete button.

#### Taint Effects

A taint can produce three possible effects:

| Effect | Description |
| --- | --- |
| **NoSchedule** | The Kubernetes scheduler will only allow scheduling pods that have tolerations for the tainted nodes. |
| **PreferNoSchedule** | The Kubernetes scheduler will try to avoid scheduling pods that do not have tolerations for the tainted nodes. |
| **NoExecute** | Kubernetes will evict the running pods from the nodes if the pods do not have tolerations for the tainted nodes. |

### Editing a Node Config

`Edit node config` allows you to directly edit any node. It will open the editor which contains all the configuration settings in which the default format is YAML. You can edit multiple objects, although changes are applied one at a time.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/clusters/edit-yaml-node.jpg)

* You can edit or modify the parameters or values of a node by clicking `Edit node config`.
* Click `Review Changes` to compare the changes in the YAML file.
* Click `Update Node`.


### Deleting a Node

* Search a node by name or label in the search bar.
* On the node, click the ellipsis button and then click `Delete`.

The node will be deleted from the cluster.

---

## Pods

---

## Other Resource Kinds

---

## Terminal

User with [Super admins](./global-configurations/authorization/user-access.md#assign-super-admin-permission) access can now troubleshoot cluster issues by accessing the cluster terminal from Devtron. You can select an image from the list that has all CLI utilities like kubectl, helm, netshoot etc. or can use a custom image, which is publicly available.

To troubleshoot a cluster or a specific node in a cluster, click the terminal symbol on the right side.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/clusters/cluster-terminal-access.jpg)

* You will see the user-defined name for the cluster in Devtron. E.g. `default-cluster`.
* In the `Nodes` field, select the node from the drop-down list you want to troubleshoot. E.g. `demo-new`.
* Select the namespace from the drop-down list which you have added in the [Environment](./global-configurations/cluster-and-environments.md#add-environment) section.
* Select the image from the drop-down list which includes all CLI utilities or you can use a custom image, which is publicly available.
* Select the terminal shell from the drop-down list (e.g. `sh`, `bash`) to troubleshoot a node.


### Example: Debug Pod 

You can also create pod for debugging which will connect to pod terminal. To find out why a particular pod is not running, you can check `Pod Events` and `Pod Manifest` for details.

* Check the current state of the Pod and recent events with the following command:

```bash
kubectl get pods
```

* To know more information about each of these pods and to debug a pod depending on the state of the pods, run the following command:

```bash
kubectl describe pod <podname>
```

| Pod Status       | Description                             | Common Causes                              |
| ---------------- | --------------------------------------- | ------------------------------------------ |
| Pending          | Pod cannot be scheduled onto a node     | Insufficient resources (CPU, memory, etc.) |
| Waiting          | Pod has been scheduled but cannot run   | Failed to pull container image             |
| CrashLoopBackOff | Container keeps crashing and restarting | Incorrect access token for API interaction |


Here, you can see configuration information about the container(s) and Pod (labels, resource requirements, etc.), as well as status information about the container(s) and Pod (state, readiness, restart count, events, etc.).


 **Note**: A container can have no shells or multiple shells running in it. If you are unable to create a successful connection, try changing the shell, as the container may not have that shell running.

---

## Creating Resources



