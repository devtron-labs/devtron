# Clusters

As `Devtron` is a tool integration platform for Kubernetes, a [cluster](https://docs.devtron.ai/getting-started#create-a-kubernetes-cluster) is created as a pre-requisite before you install Devtron depending on your [resource usage and requirements](https://docs.devtron.ai/getting-started#recommended-resources).

By integrating into a Kubernetes cluster, Devtron helps you to deploy, observe, manage and debug the existing Helm apps in all your clusters.

On the left navigation of Devtron, select `Clusters`. You will find the list of clusters in this section which you have added under [Global Configurations > Clusters & Environments](https://docs.devtron.ai/global-configurations/cluster-and-environments).

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/clusters/clusters.jpg)


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

## Resources

On the `Resource allocation and usage`, you can see the details of compute resources or resources.

* CPU resources
* Memory resources

If you specify a `request` and `limits` in the container resource manifest file, then the respective values will appear on the `Resource allocation and usage` section.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/clusters/resource.jpg)

## Nodes

You can see the list of nodes available in your cluster. Typically you have several nodes in a cluster; in a learning or resource-limited environment, you might have only one node.

The components on a typical node include the `kubelet`, a `container runtime`, and the `kube-proxy`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/clusters/nodes.jpg)

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

To see the summary of a node, click the particular node.

## Troubleshoot Cluster via Terminal

User with [Super admins](https://docs.devtron.ai/global-configurations/authorization/user-access#assign-super-admin-permissions) access can now troubleshoot cluster issues by accessing the cluster terminal from Devtron. You can select an image from the list that has all CLI utilities like kubectl, helm, netshoot etc. or can use a custom image.

To troubleshoot a cluster or a specific node in a cluster, click the terminal symbol on the right side.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/clusters/cluster-terminal-access.jpg)

* You will see the user-defined name for the cluster in Devtron. E.g. `default-cluster`.
* In the `Nodes` field, select the node from the drop-down list you want to troubleshoot. E.g. `demo-new`.
* Select the namespace from the drop-down list which you have added in the [Environment](https://docs.devtron.ai/global-configurations/cluster-and-environments#add-environment) section.
* Select the image from the drop-down list which includes all CLI utilities or you can use a custom image.
* Select the terminal shell from the drop-down list (e.g. `sh`, `bash`, `powershell`, `cmd`) to troubleshoot a node.


 **Note**: A pod can have one or more containers running, and a container can have no or multiple shells running in it. If you are not able to create a successfull connection, try changing the shell, as the container may not have that shell running.


