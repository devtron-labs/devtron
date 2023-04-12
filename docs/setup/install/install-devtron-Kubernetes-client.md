# Devtron Kubernetes Client

## Overview

The Kubernetes client by Devtron is a very lightweight dashboard that can be installed on arm64/amd64-based architectures. It comes with the features such as Kubernetes Resources Browser and Cluster Management that can provide control and observability for resources across clouds and clusters.

Devtron Kubernetes client is an intuitive Kubernetes Dashboard or a command line utility installed outside a Kubernetes cluster. The client can be installed on a desktop running on any Operating Systems and interact with all your Kubernetes clusters and workloads through an API server. It is a binary, packaged in a bash script that you can download and install by using the following set of commands.

By installing `Devtron Kubernetes Client`, you can access:

* [Kubernetes Resource Browser](#kubernetes-resource-browser)
* [Clusters Management Feature](#cluster-management)


## Here are a few advantages of using Devtron Kubernetes Client:

* **Managing Kubernetes Resources at scale**: Clusters vary on business and architectural needs. Organizations tend to build smaller clusters for more decentralization. This practice leads to the creation of multiple clusters and more nodes. Managing them on a CLI requires multiple files, making it difficult to perform resource operations. But with the Devtron Kubernetes Client, you can gain more visibility into K8s resources easily.

* **Unifying information in one place**: When information is scattered across clusters, and you have to type commands with arguments to fetch desired output, the process becomes slow and error-prone. Without a single point of configuration source, the configurations of different config. files diverge, making them even more challenging to restore and track. The Devtron Kubernetes Client unifies all the information and tools into one interface to perform various contextual tasks.

* **Accessibility during an outage for troubleshooting**: As the Devtron Kubernetes Client runs outside a cluster, you can exercise basic control over their failed resources when there is a cluster-level outage. The Client helps to gather essential logs and data to pinpoint the root cause of the issue and reduce the time to restore service.

* **Avoiding Kubeconfig version mismatch errors**: With the Devtron Kubernetes Client, you can be relieved from maintaining the Kubeconfig versions for the respective clusters (v1.16 - 1.26 i.e, current version) as the Devtron Kubernetes Client performs self kubeconfig version control. Instead of managing multiple kubectl versions manually, it eliminates the chances of errors occurring due to the mismatch in configuration. 


## Install Devtron Kubernetes Client

* Download the bash script using the below URL:
https://cdn.devtron.ai/k8s-client/devtron-install.bash

* To automatically download the executable file and to open the dashboard in the respective browser, run the following command:

```bash
   sh devtron-install.bash start  
```
`Note`: Make sure you place `Devtron-install.bash` in your current directory before you execute the command.

* Devtron Kubernetes Client opens in your browser automatically.

* You must add your cluster to make your cluster visible on the `Kubernetes Resource Browser` and `Clusters` section. To add a cluster, go to the `Global Configurations` and click `Add Cluster`. [Refer documentation on how to add a cluster](https://docs.devtron.ai/v/v0.6/global-configurations/cluster-and-environments#add-cluster).

`Note`: You do not need to have a `super admin` permission to add a cluster if you install `Devtron Kubernetes Client`. You can add more than one cluster.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/install+devtron+K8s+client/global-configs-clusters.jpg)


### Kubernetes Resource Browser

`Kubernetes Resource Browser` provides a graphical user interface for interacting and managing all your Kubernetes (k8s) resources across clusters. It also helps you to deploy and manage Kubernetes resources and allows pod operations such as:
* View real-time logs
* Check manifest and edit live manifests of k8s resources
* Executable via terminal
* View Events
* Or, delete a resource

With `Kubernetes Resource browser`, you can also perform the following:
* Check the real-time health status
* Search for any workloads
* Manage multiple clusters and change cluster contexts
* Deploy multiple K8s manifests through `Create` UI option.
* Perform resource grouping at the cluster level.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/install+devtron+K8s+client/k8s-resource-browser.jpg)

After your cluster is added via `Global Configurations`, go to the `Kubernetes Resource Browser` page and select your cluster. [Refer Resource Browser documentation for detail and its operations](https://docs.devtron.ai/v/v0.6/usage/resource-browser).

`Note`: You do not need to have a `super admin` permission to access `Kubernetes Resource Browser` if you install `Devtron Kubernetes Client`.


### Cluster Management

With the `Devtron Kubernetes Client`, you can manage all your clusters running on-premises or on a cloud. It is a cluster and cloud agnostic platform where you can add as many clusters as you want, be it a lightweight cluster such as k3s/ microk8s or cloud managed clusters like Amazon EKS. 

It enables you to observe and monitor the cluster health and real-time node conditions. The Cluster management feature provides a summary of nodes with all available labels, annotations, taints, and other parameters such as resource usage. In addition to that, it helps you to perform node operations such as:

* Debug a node
* Cordon a node
* Drain a node
* Taint a node
* Edit a node config
* Delete a node

With its rich features and intuitive interface, you can easily manage and [debug clusters through cluster terminal access](https://docs.devtron.ai/v/v0.6/usage/clusters#access-cluster-via-terminal-for-troubleshooting) and use any CLI debugging tools like busybox, kubectl, netshoot or any custom CLI tools like k9s.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/install-devtron/install+devtron+K8s+client/cluster-terminal.jpg)


After your cluster is added via `Global Configurations`, go to the `Clusters` page and search or select your cluster. [Refer Clusters documentation for detail and its operations](https://docs.devtron.ai/v/v0.6/usage/clusters).


### Some Peripheral Commands

* In case if you close the browser by mistake, you can open the dashboard by executing the following command. It will open the dashboard through a port in the available web browser and store the Kubernetes client's state.

```bash
sh devtron-install.bash open 
```

* To stop the dashboard, you can execute the following command:

```bash
sh devtron-install.bash stop
``` 

* To update the `Devtron Kubernetes Client`, use the following command. It will stop the running dashboard and download the latest executable file and open it in the browser.

```bash
sh devtron-install.bash upgrade
```



