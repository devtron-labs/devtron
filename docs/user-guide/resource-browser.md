# Resource Browser

## Introduction

The Devtron Resource Browser provides you a central interface to view and manage all your [Kubernetes objects](../../reference/glossary.md#objects) across clusters.  It helps you perform key actions like viewing logs, editing live manifests, and even creating/deleting resources directly from the user interface. This is especially useful for troubleshooting purposes as it supports multi-cluster too.

First, the Resource Browser shows you a list of clusters added to your Devtron setup. By default, it displays a cluster named '*default_cluster*' after the [initial setup](../../setup/install/README.md) is successful.

![Figure 1: Devtron Resource Browser - List of Clusters](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/resource-browser.jpg)

You can easily connect more clusters by clicking the **Add Cluster** button located at the top of the browser. This will take you to the [Cluster & Environments](../global-configurations/cluster-and-environments.md) configuration within [Global Configurations](../global-configurations/README.md).

You may click a cluster to view and manage all its resources as shown below.

![Figure 2: Resources within Cluster](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/resource-list.jpg)

---

## Overview

![Figure 3: Resource Browser - Overview Page](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/rb-overview.jpg)

### Resource Utilization

This shows the combined CPU and memory consumption of all running pods in the cluster.

| Parameter       | Description                                                                                                 |
| --------------- | ------------------------------------------------------------------------------------------------------------|
| CPU Usage       | Percentage of CPU resources currently being used across all the pods in the cluster.                        |
| CPU Capacity   | Total amount of CPU resources available across all the nodes in the cluster. Measured in millicores (m).    |
| CPU Requests   | Total amount of CPU resources requested by all the pods in the cluster.                                     |
| CPU Limits      | Maximum amount of CPU resources that a pod can use.                                                         |
| Memory Usage   | Percentage of memory resources currently being used across all the pods in the cluster.                     |
| Memory Capacity | Total amount of memory resources available across all the nodes in the cluster. Measured in Megabytes (Mi). |
| Memory Requests | Total amount of memory resources requested by all the pods in the cluster.                                  |
| Memory Limits  | Maximum amount of memory resources that a pod can use.                                                      |

### Errors

This shows errors in the cluster. If no error is present in the cluster, Resource Browser will not display this section.

### Catalog Framework

Based on the schema provided in the catalog framework, you can add relevant details for each cluster. Refer [Catalog Framework](./global-configurations/catalog-framework.md) for more details.

### Readme

You can also include additional information about your cluster using the Markdown editor.

---

## Discovering Resources 

### Search and Filter

You can use the searchbox to browse the resources.

![Figure 4: Locate Resources using Searchbox](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/discover-resource.gif)

Moreover, you can use filters that allow you to quickly filter your workload as per labels, field selectors, or [CEL expression](https://kubernetes.io/docs/reference/using-api/cel/) as shown below.

{% embed url="https://www.youtube.com/watch?v=E-V-ELCXtfs" caption="Filtering Workloads in Devtron" %}

### Edit Manifest 

You can edit the [manifest](../reference/glossary.md#manifest) of a Kubernetes object. This can be for fixing errors, scaling resources, or changing configuration.

![Figure 5: Editing a Live Manifest](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/edit-live-manifest.gif)

### View Events

You can monitor activities like creation, deletion, updation, scaling, or errors in the resources involved.

![Figure 6: Viewing All Events](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/events.gif)

### Delete

You can delete an unwanted resource if it is orphaned and no longer required by your applications.

![Figure 7: Deleting a Resource](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/delete.gif)

You can perform all the actions on the following:
* [Nodes](#nodes)
* [Pods](#pods)
* [Other Resource Kinds](#other-resource-kinds)

---

## Nodes

You can see the list of nodes available in your cluster. Typically you have several nodes in a cluster; in a learning or resource-limited environment, you might have only one node.

The components on a typical node include the `kubelet`, a `container runtime`, and the `kube-proxy`.

If you have multiple nodes, you can search a node by name or label in the search bar. The search result will display the following information of the node. To display a parameter of a node, use `Columns` on the right side, select the parameter to display from the drop-down list, and click **Apply**.

![Figure 8: Searching and Filtering Nodes](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/cluster-nodes.jpg)

| Fields | Description |
| --- | --- |
| Node | User-defined name for the node in Devtron. E.g. `demo-new`.<br>Note: Two nodes cannot have the same name at the same time.</br> |
| Status | Status of a node. It can be either `Ready` or `Not Ready`. |
| Roles | Shows the roles of a node. |
| Errors | Shows the error in nodes. |
| K8s Version | Shows the version of Kubernetes cluster. |
| Node Group | Shows which collection of worker nodes it belongs to. |

Clicking on a node shows you a number of details such as:

* CPU Usage and Memory Usage of Node
* CPU Usage and Memory Usage of Each Pod
* Number of Pods in the Node
* List of Pods
* Age of Pods
* Labels, Annotations, and Taints
* Node IP

![Figure 9: Checking Node Summary](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/node-summary.jpg)

Further using the Devtron UI, you will be able to:
* [Debug a Node](#debug-a-node)
* [Cordon a Node](#cordon-a-node)
* [Drain a Node](#drain-a-node)
* [Taint a Node](#taint-a-node)
* [Edit a Node Config](#edit-a-node-config)
* [Delete a Node](#delete-a-node)

{% hint style="info" %}
### Why are node operations required?
Your applications run on pods, and pods run on nodes. But sometimes, Kubernetes scheduler cannot deploy a pod on a node for several reasons, e.g., node is not ready, node is not reachable, network is unavailable, etc. In such cases, node operations help you manage the nodes better.
{% endhint %}

### Debug a Node

You can debug a node via [Cluster Terminal](#cluster-terminal) by selecting your namespace and image from the list that has all CLI utilities like kubectl, helm, netshoot etc. or can use a custom image, which is publicly available.

* Click **Debug**.

  ![Figure 10a: Debugging a Node](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/debug.jpg)

* Debug a node by selecting the terminal shell, i.e., `bash` or `sh`.

  ![Figure 10b: Debug Terminal](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/debug-terminal.jpg)

### Cordon a Node

Cordoning a node means making the node unschedulable. After cordoning a node, new pods cannot be scheduled on this node.

![Figure 11a: Visual Representation - Cordoning a Node](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/cordon-visual.jpg)

* Click **Cordon**.

  ![Figure 11b: Cordoning a Node](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/cordon.jpg)

* A confirmation dialog box will appear, click **Cordon Node** to proceed.

  ![Figure 11c: Cordon Confirmation](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/cordon-dialog.jpg)

The status of the node shows `SchedulingDisabled` with `Unschedulable` parameter set as `true`.

Similarly, you can uncordon a node by clicking `Uncordon`. After a node is uncordoned, new pods can be scheduled on the node.

### Drain a Node

Before performing maintenance on a node, draining a node evicts all of your pods safely from a node. Safe evictions allow the pod’s containers to gracefully terminate and honour the `PodDisruptionBudgets` you have specified (if relevant).

![Figure 12a: Visual Representation - Draining a Node](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/drain-visual.jpg)

After the node is drained, all pods (including those managed by DaemonSets) in the node will be automatically drained to other nodes in the cluster, and the drained node will be set to cordoned status.

* Click **Drain**.

  ![Figure 12b: Draining a Node](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/drain.jpg)

* A confirmation dialog box will appear, click **Drain Node** to proceed.

  ![Figure 12c: Drain Confirmation](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/drain-dialog.jpg)

You can also select from the following conditions before draining a node:

| Name | Usage |
| --- | --- |
| Grace Period | Period of time in seconds given to each pod to terminate gracefully. If negative, the default value specified in the pod will be used. |
| Delete empty directory data | Enabling this field will delete the pods using empty directory data when the node is drained. |
| Disable eviction (use with caution) | Enabling this field will force drain to use delete, even if eviction is supported. This will bypass checking `PodDisruptionBudgets`.<br>Note: Make sure to use with caution.</br> |
| Force drain | Enabling this field will force drain a node even if there are pods that do not declare a controller. |
| Ignore DaemonSets | Enabling this field will ignore DaemonSet-managed pods. |

### Taint a Node

Taints are `key:value` pairs associated with effect. After you add taints to nodes, you can set tolerations on a pod to allow the pod to be scheduled to nodes with certain taints. When you taint a node, it will repel all the pods except those that have a toleration for that taint. A node can have one or many taints associated with it.

![Figure 13a: Visual Representation - Tainting a Node](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/taint-visual.jpg)

**Note**: Make sure to check taint validations before you add a taint.

* Click **Edit taints**.

  ![Figure 13b: Tainting a Node](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/edit-taints.jpg)

* Enter the `key:value` pairs and select the [taint effect](#taint-effects) from the drop-down list.

  ![Figure 13c: Adding Taints](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/taint-dialog.jpg)

* Click **Save**.

You can also add more taints using **+ Add taint button**, or delete the existing taint by using the delete icon.

#### Taint Effects

A taint can produce three possible effects:

| Effect | Description |
| --- | --- |
| NoSchedule | The Kubernetes scheduler will only allow scheduling pods that have tolerations for the tainted nodes. |
| PreferNoSchedule | The Kubernetes scheduler will try to avoid scheduling pods that do not have tolerations for the tainted nodes. |
| NoExecute | Kubernetes will evict the running pods from the nodes if the pods do not have tolerations for the tainted nodes. |

### Edit a Node Config

`Edit node config` allows you to directly edit any node. It will open the editor which contains all the configuration settings in which the default format is YAML. You can edit multiple objects, although changes are applied one at a time.

![Figure 14: Editing Node Config](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/edit-config.gif)

* You can edit or modify the parameters or values of a node by clicking **Edit node config**.
* Click **Review Changes** to compare the changes in the YAML file.
* Click **Update Node**.


### Delete a Node

* Search a node by name or label in the search bar.
* On the node, click the ellipsis button and then click **Delete**.

![Figure 15a: Deleting a Node](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/delete-node.jpg)

![Figure 15b: Delete Confirmation](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/delete-dialog.jpg)

The node will be deleted from the cluster.

{% hint style="info" %}
You can also access [Cluster Terminal](#cluster-terminal) from your node.
{% endhint %}


---

## Pods

### Manifest

Shows you the [configuration](../reference/glossary.md#manifest) of the selected pod and allows you to edit it. Refer [Edit Manifest](#edit-manifest) to know more.

### Events

Shows you all the activities (create/update/delete) of the selected pod. Refer [View Events](#view-events) to know more.

### Logs

Examining your cluster's pods helps you understand the health of your application. By inspecting pod logs, you can check the performance and identify if there are any failures. This is especially useful for debugging any issues effectively.

Moreover, you can download the pod logs for ease of sharing and troubleshooting as shown in the below video.

{% embed url="https://www.youtube.com/watch?v=PP0ZKAZCT58" caption="Downloading Pod Logs" %}

#### Last Restart Pod Log

In case any of your pod restarts, you can determine the cause by viewing its container log, time, status, node events in the pod listing screen as shown below:

![Figure 16: Checking Restart Pod Log](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/restart-pod-log.gif)

### Terminal

You can access the terminal within a running container of a pod to view its logs, troubleshoot issues, or execute commands directly. This is different from the [cluster terminal](#cluster-terminal) you get at node level. 

#### Launching Ephemeral Container

This is a part of [Pod Terminal](#pod-terminal). It is especially useful when `kubectl exec` is insufficient because a container has crashed or a container image doesn't include debugging utilities.

{% embed url="https://www.youtube.com/watch?v=Ml19i29Ivc4" caption="Launching Ephemeral Containers from Resource Browser" %}

1. In the Resource Browser, select **Pod** within `Workloads`.
2. Use the searchbar to find and locate the pod you wish to debug. Click the pod.
3. Go to the **Terminal** tab 
4. Click **Launch Ephemeral Container** as shown below.

    You get 2 tabs:
    1. **Basic** - It provides the bare minimum configurations required to launch an ephemeral container.

    ![Figure 17: Basic Tab](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/debugging-deployment-and-monitoring/basic.jpg)

    It contains 3 mandatory fields:

    * **Container name prefix** - Type a prefix to give to your ephemeral container, for e.g., *debug*. Your container name would look like `debug-jndvs`.

    * **Image** - Choose an image to run from the dropdown. Ephemeral containers need an image to run and provide the capability to debug, such as `curl`. You can use a custom image too.
    
    * **Target Container name** - Since a pod can have one or more containers, choose a target container you wish to debug, from the dropdown.

    2. **Advanced** - It is particularly useful for advanced users that wish to use labels or annotations since it provides additional key-value options. Refer [Ephemeral Container Spec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#ephemeralcontainer-v1-core) to view the supported options.
    
    {% hint style="info" %}
    Devtron ignores the 'command' field while launching an ephemeral container
    {% endhint %}

---

## Other Resource Kinds

Other resources in the cluster are grouped under the following categories:

* Namespace
* Workloads
* Config & Storage
* Networking
* RBAC
* Administration
* Other Resources
* Custom Resource

![Figure 18: Resources within Cluster](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/resource-list.jpg)

---

## Cluster Terminal

User with [super-admin](./global-configurations/authorization/user-access.md#assign-super-admin-permission) access can now troubleshoot cluster issues by accessing the cluster terminal from Devtron. You can select an image from the list that has all CLI utilities like kubectl, helm, netshoot etc. or can use a custom image, which is publicly available.

To troubleshoot a cluster or a specific node in a cluster, click the terminal icon on the right side.

![Figure 19: Terminal Icon](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/cluster-terminal.gif)

* You will see the user-defined name for the cluster in Devtron. E.g. `default-cluster`.
* Select the node you wish to troubleshoot from the `Node` drop-down. E.g. `demo-new`.
* Select the namespace from the drop-down list which you have added in the [Environment](./global-configurations/cluster-and-environments.md#add-environment) section.
* Select the image from the drop-down list which includes all CLI utilities or you can use a custom image, which is publicly available.
* Select the terminal shell from the drop-down list (e.g. `sh`, `bash`) to troubleshoot a node.

### Use Case - Debugging Pods

You can also create pod for debugging which will connect to pod terminal. To find out why a particular pod is not running, you can check `Pod Events` and `Pod Manifest` for details.

The **Debug Mode** is helpful in scenarios where you can't access your Node by using an SSH connection. When enabled, a pod is created on the node, which opens an interactive shell on the Node.

* Check the current state of the pod and recent events with the following command:

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

Here, you can see configuration information about the container(s) and pod (labels, resource requirements, etc.), as well as status information about the container(s) and pod (state, readiness, restart count, events, etc.).

{% hint style="info" %}
A container can have no shells or multiple shells running in it. If you are unable to create a successful connection, try changing the shell, as the container may not have that shell running.
{% endhint %}

---

## Creating Resources

You can create one or more [Kubernetes objects](../reference/glossary.md#objects) in your cluster using YAML. In case you wish to create multiple objects, separate each resource definition by three dashes (---).

Once you select a cluster in Resource Browser, click **+ Create Resource**, and add the resource definition.  

In the below example, we have created a simple pod named `nginx`:

![Figure 20: Creating Resources within Cluster](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/create-resource.gif)

Here's one more example that shows the required fields and object specifications for a Kubernetes Deployment:

{% code title="Spec File" overflow="wrap" lineNumbers="true" %}

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels: 
     app: nginx
spec:
  replicas: 2
  selector:
    matchLabels:
       app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
       - name: nginx
         image: nginx:1.14.2
         ports:
         - containerPort: 80
```
{% endcode %}
