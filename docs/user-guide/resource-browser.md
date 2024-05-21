# Resource Browser

## Introduction

The Devtron Resource Browser provides you a central interface to view and manage all your [Kubernetes objects](../../reference/glossary.md#objects) across clusters.  It helps you perform key actions like viewing logs, editing live manifests, and even creating/deleting resources directly from the user interface. This is especially useful for troubleshooting purposes as it supports multi-cluster too.

{% hint style="info" %}
### Additional References
* [Resource browser versus traditional tools like kubectl](https://devtron.ai/blog/managing-kubernetes-resources-across-multiple-clusters)
* [Why you should use Devtron's Resource Browser](https://devtron.ai/blog/what-is-the-kubernetes-resource-browser-in-devtron)
{% endhint %}

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
| CPU Limits      | Maximum amount of CPU resources that a total number of pods can use in the cluster.                        |
| Memory Usage   | Percentage of memory resources currently being used across all the pods in the cluster.                     |
| Memory Capacity | Total amount of memory resources available across all the nodes in the cluster. Measured in Megabytes (Mi). |
| Memory Requests | Total amount of memory resources requested by all the pods in the cluster.                                  |
| Memory Limits  | Maximum amount of memory resources that a total number of pods can use in the cluster.                       |

### Errors

This shows errors in the cluster. If no error is present in the cluster, Resource Browser will not display this section.

### Catalog Framework

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to edit the catalog framework.
{% endhint %}

Based on the schema provided in the catalog framework, you can add relevant details for each cluster. Refer [Catalog Framework](./global-configurations/catalog-framework.md) for more details. 

### Readme

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have super-admin permission to edit the readme file.
{% endhint %}

You can also include additional information about your cluster using the Markdown editor.

---

## Discovering Resources 

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have [access to the cluster](./global-configurations/authorization/user-access.md#kubernetes-resources-permissions) to discover resources.
{% endhint %}

### Search and Filter

You can use the searchbox to browse the resources.

![Figure 4: Locate Resources using Searchbox](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/discover-resource.gif)

Moreover, you can use filters that allow you to quickly filter your workload as per labels, field selectors, or [CEL expression](https://kubernetes.io/docs/reference/using-api/cel/) as shown below.

{% embed url="https://www.youtube.com/watch?v=E-V-ELCXtfs" caption="Filtering Workloads in Devtron" %}

### Edit Manifest 

{% hint style="warning" %}
### Who Can Perform This Action?
User needs to be an [admin of the Kubernetes resource](./global-configurations/authorization/user-access.md#kubernetes-resources-permissions) to edit its manifest.
{% endhint %}

You can edit the [manifest](../reference/glossary.md#manifest) of a Kubernetes object. This can be for fixing errors, scaling resources, or changing configuration.

![Figure 5: Editing a Live Manifest](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/edit-live-manifest.gif)

### View Events

You can monitor activities like creation, deletion, updation, scaling, or errors in the resources involved. Refer [Events](https://kubernetes.io/docs/reference/kubernetes-api/cluster-resources/event-v1/) to learn more.

![Figure 6: Viewing All Events](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/events.gif)

### Delete

{% hint style="warning" %}
### Who Can Perform This Action?
User needs to be an [admin of the Kubernetes resource](./global-configurations/authorization/user-access.md#kubernetes-resources-permissions) to delete it.
{% endhint %}

You can delete an unwanted resource if it is orphaned and no longer required by your applications.

![Figure 7: Deleting a Resource](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/delete.gif)

---

## Nodes

You can see the list of nodes available in your cluster. Typically you have several nodes in a cluster; in a learning or resource-limited environment, you might have only one node.

The components on a typical node include the `kubelet`, a `container runtime`, and the `kube-proxy`.

If you have multiple nodes, you can search a node by name or label in the search bar. The search result will display the following information about the node. To display a parameter of a node, use `Columns` on the right side, select the parameter to display from the drop-down list, and click **Apply**.

![Figure 8: Searching and Filtering Nodes](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/cluster-nodes.jpg)

| Fields | Description |
| --- | --- |
| Node | Alphanumeric name of the node |
| Status | Status of a node. It can be either `Ready` or `Not Ready`. |
| Roles | Shows the roles of a node, e.g., agent |
| Errors | Shows the number of errors in nodes (if any) |
| K8s Version | Shows the version of Kubernetes cluster |
| Node Group | Shows which collection of worker nodes it belongs to |
| No. of Pods | Shows the total number of pods present in the node |

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
### Why Are Node Operations Required?
Your applications run on pods, and pods run on nodes. But sometimes, Kubernetes scheduler cannot deploy a pod on a node for several reasons, e.g., node is not ready, node is not reachable, network is unavailable, etc. In such cases, node operations help you manage the nodes better.
{% endhint %}

{% hint style="warning" %}
### Who Can Perform These Actions?
Users need to have super-admin permission to perform node operations.
{% endhint %}

### Debug a Node

You can debug a node via [Cluster Terminal](#cluster-terminal) by selecting your namespace and image from the list that has all CLI utilities like kubectl, helm, netshoot etc. or can use a custom image, which is publicly available.

* Click **Debug**.

  ![Figure 10a: Debugging a Node](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/debug.jpg)

* Debug a node by selecting the terminal shell, i.e., `bash` or `sh`.

  ![Figure 10b: Debug Terminal](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/debug-terminal.jpg)

### Cordon a Node

Cordoning a node means making the node unschedulable. After [cordoning a node](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_cordon/), new pods cannot be scheduled on this node.

![Figure 11a: Visual Representation - Cordoning a Node](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/cordon-visual.jpg)

* Click **Cordon**.

  ![Figure 11b: Cordoning a Node](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/cordon.jpg)

* A confirmation dialog box will appear, click **Cordon Node** to proceed.

  ![Figure 11c: Cordon Confirmation](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/cordon-dialog.jpg)

The status of the node shows `SchedulingDisabled` with `Unschedulable` parameter set as `true`.

Similarly, you can uncordon a node by clicking `Uncordon`. After a node is uncordoned, new pods can be scheduled on the node.

### Drain a Node

Before performing maintenance on a node, [draining a node](https://kubernetes.io/docs/tasks/administer-cluster/safely-drain-node/) evicts all of your pods safely from a node. Safe evictions allow the pod’s containers to gracefully terminate and honour the `PodDisruptionBudgets` you have specified (if relevant).

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

{% hint style="info" %}
### Additional Reference
[Click here](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/#concepts) to read about taint effects.
{% endhint %}

### Edit a Node Config

This allows you to directly edit any node. It will open the editor which contains all the configuration settings in which the default format is YAML. You can edit multiple objects, although changes are applied one at a time.

![Figure 14: Editing Node Config](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/edit-config.gif)

* Go to the `YAML` tab and click **Edit YAML**.
* Make the changes using the editor.
* Click **Review & Save changes** to compare the changes in the YAML file.
* Click **Apply changes** to confirm.


### Delete a Node

You can also delete a node by clicking the **Delete** button present on the right-hand side.

![Figure 15a: Deleting a Node](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/delete-node.jpg)

![Figure 15b: Delete Confirmation](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/delete-dialog.jpg)

The node will be deleted from the cluster.

{% hint style="info" %}
You can also access [Cluster Terminal](#cluster-terminal) from your node.
{% endhint %}

---

## Pods

{% hint style="warning" %}
### Who Can Perform This Action?
Users need to have [access to the cluster](./global-configurations/authorization/user-access.md#kubernetes-resources-permissions) to view its pods and its data.
{% endhint %}

### Manifest

Shows you the [configuration](../reference/glossary.md#manifest) of the selected pod and allows you to edit it. Refer [Edit Manifest](#edit-manifest) to learn more.

### Events

Shows you all the activities (create/update/delete) of the selected pod. Refer [View Events](#view-events) to know more.

### Logs

Examining your cluster's pods helps you understand the health of your application. By inspecting pod logs, you can check the performance and identify if there are any failures. This is especially useful for debugging any issues effectively.

Moreover, you can download the pod logs for ease of sharing and troubleshooting as shown in the below video.

{% embed url="https://www.youtube.com/watch?v=PP0ZKAZCT58" caption="Downloading Pod Logs" %}

#### Pod Last Restart Snapshot

Frequent pod restarts can impact your application as it might lead to unexpected downtimes. In such cases, it is important to determine the root cause and take actions (both preventive and corrective) if needed.

In case any of your pod restarts, you can view its details from the pod listing screen:
* Last pod restart event, along with the timestamp and message
* Reason behind restart
* Container log before restart
* Node status and events  

![Figure 16: Checking Restart Pod Log](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/restart-pod-log.gif)

### Terminal

{% hint style="warning" %}
### Who Can Perform This Action?
User needs to be an [admin of the Kubernetes resource](./global-configurations/authorization/user-access.md#kubernetes-resources-permissions) to access pod terminal.
{% endhint %}

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

You can also create a pod for debugging which will connect to the pod terminal. To find out why a particular pod is not running, you can check `Pod Events` and `Pod Manifest` for details.

The **Auto select** option automatically selects a node from a list of nodes and then creates a pod. Alternatively, you can choose a node of your choice from the same dropdown for debugging.

The **Debug Mode** is helpful in scenarios where you can't access your node by using an SSH connection. Enabling this feature opens an interactive shell directly on the node. This shell provides unrestricted access to the node, giving you enhanced debugging capabilities.

* Check the current state of the pod and recent events with the following command:

```bash
kubectl get pods
```

* To know more information about each of these pods and to debug a pod depending on the state of the pods, run the following command:

```bash
kubectl describe pod <podname>
```

Here, you can see configuration information about the container(s) and pod (labels, resource requirements, etc.), as well as status information about the container(s) and pod (state, readiness, restart count, events, etc.). [Click here](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/) to know more about pod lifecycle.

{% hint style="info" %}
A container can have no shells or multiple shells running in it. If you are unable to create a successful connection, try changing the shell, as the container may not have that shell running.
{% endhint %}

---

## Creating Resources

{% hint style="warning" %}
### Who Can Perform This Action?
User needs to be an [admin of the Kubernetes resources](./global-configurations/authorization/user-access.md#kubernetes-resources-permissions) to create resources.
{% endhint %}

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

---

## Port Forwarding

### Introduction

Assume your applications are running in a Kubernetes cluster on cloud. Now, if you wish to test or debug them on your local machine, you can perform [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/). It creates a tunnel between a port on your machine and a port on a resource within your cluster. Therefore, you can access applications running inside the cluster as though they are running locally on your machine.

But first, you would need access to that cluster. Traditionally, the kubeconfig file (`./kube/config`) helps you connect with the cluster. 

![Figure 21: Kubeconfig File](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/kubeconfig.jpg)

### Challenges in Kubeconfig

Kubeconfig becomes painstakingly difficult to maintain especially when it comes to:
* Granting or revoking access to the cluster for multiple people
* Changing the permissions and subsequently the access token
* Adding/Updating/Deleting the entries of cluster URLs and tokens
* Keeping a record of multiple kubeconfig files

### Our Solution

Devtron helps in reducing the challenges and simplifying the maintenance of kubeconfig file through:
* **Devtron's Proxy URL for Cluster** - A standardized URL that you can use in place of your Kubernetes cluster URL.
* **Devtron's Access Token** - A kubectl-compatible token which can be generated and centrally maintained from [Global Configurations → Authorization → API tokens](./global-configurations/authorization/api-tokens.md).

### Steps

**Prerequisite**: An [API token with necessary permissions](./global-configurations/authorization/api-tokens.md) for the user(s) to access the cluster. 

If you are not a super-admin and can't generate a token yourself, you can find the session token (argocd.token) using the Developer Tools available in your web browser as shown below.

![Figure 22: Using Session Token](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/argocd-token-v1.gif)

1. Go to `~/.kube` folder on your local machine and open the `config` file. Or you may create one with the following content:

  {% code title="kubeconfig" overflow="wrap" lineNumbers="true" %}
  ```yaml
  apiVersion: v1
  kind: Config
  clusters:
  - cluster:
      insecure-skip-tls-verify: true
      server: https://<devtron_host_name>/orchestrator/k8s/proxy/cluster/<cluster_name>
    name: devtron-cluster
  contexts:
  - context:
      cluster: devtron-cluster
      user: admin
    name: devtron-cluster
  current-context: devtron-cluster
  users:
  - name: admin
    user:
      token: <devtron_token>
  ```
  {% endcode %}

2. Edit the following placeholders in the `server` field and the `token` field:

  | Placeholder         | Description                         | Example                                         |
  | ------------------- | ----------------------------------- | ----------------------------------------------- |
  | <devtron_host_name> | Hostname of the Devtron server      | demo.devtron.ai                                 |
  | <cluster_name>      | Name of the cluster (or cluster ID) | devtron-cluster                                 |
  | <devtron_token>     | API token or session token          | \-                                              |

  ![Figure 23: Editing Kubeconfig File](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/kubeconfig.gif)

3. Test the connection to the cluster by running any kubectl command, e.g., `kubectl get ns` or `kubectl get po -A`

4. Once you have successfully connected to the cluster, you may run the port-forward command. Refer [kubectl port-forward](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_port-forward/) to see a few examples.

  ![Figure 24: Example - Port Forwarding](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/port-forward.gif)

