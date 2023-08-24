# Resource Browser

`Resource Browser` lists all of the resources running in your cluster. You can use it to view, inspect, manage, and delete resources in your cluster. You can also create resources from the `Resource Browser`.

Resource Browser are helpful for troubleshooting issues. It supports multi-cluster.

**Note**: To provide permission to a user to view, inspect, manage, and delete resources, go to the [Authorization > User Permissions](https://docs.devtron.ai/global-configurations/authorization/user-access) section of `Global Configurations`. You can also provide permission via [API token](https://docs.devtron.ai/global-configurations/authorization/api-tokens) or [Permission groups](https://docs.devtron.ai/global-configurations/authorization/permission-groups). Only super admin users will be able to see `Kubernetes Resources` tab and provide permission to other users to access `Resource Browser`.

Please also note that `Resource Browser` page is under the early version of development and its a `Beta` release.

## Kubernetes Resources 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/kubernetes-resource-browser-latest.jpg)

The following resources are grouped in the categories:

* **Workloads** displays workloads (Cronjob, Deployment, StatefulSet, DaemonSet, Job, and Pod resources) deployed to clusters in your current project. Includes each workload's name, status, type, number of running and total desired Pods, namespace, and cluster.

* **Config & Storage** display ConfigMap, Secret, PersistantVolume, PersistentVolumeClaim, Pod DisruptionBudget, resources which are used by applications for storing data. The configMap and secret data are provided as local ephemeral storage, which means there is no long-term guarantee about durability. A PersistentVolume (PV) is a piece of storage in the cluster that has been provisioned by server/storage/cluster administrator or dynamically provisioned using Storage Classes. It is a resource in the cluster just like node. Whereas, A PersistentVolumeClaim (PVC) is a request for storage by a user which can be attained from PV. It is similar to a Pod. Pods consume node resources and PVCs consume PV resources.


* **Networking** displays your project's [Endpoints](https://kubernetes.io/docs/concepts/services-networking/endpoint-slices/), [NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/) and [Service](https://kubernetes.io/docs/concepts/services-networking/service/) resources. Displays each resource's name, namespace, type, Cluster-IP, External-IP, Port(s), Age.

* **RBAC** stands for Role-based access control which provides the authorization strategy for regulating access to cluster or resources based on the roles of individual users within your organization.

* **Custom Resource** allows you to create your own API resources and define your own kind just like Pod, Deployment, ReplicaSet, etc. 


The following resources are grouped as uncategoried:

** **Events** displays all the reports of an event in a cluster.

** **Namespaces** displays the current list of namespaces in a cluster.

### Search and Filter Resources

You can search and filter resources by specific resource Kinds. You can also preview `Manifest`, `Events`, `Logs`, access `Terminal` by selecting ellipsis on the specific resource or `Delete` a specific resource.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/select-resource.jpg)

#### Manifest


A manifest is a YAML file that describes each component or resource of your deployment and the state you want your cluster to be in once applied. Once you deploy, you can always edit your manifest file. A manifest specifies the desired state of an object that Kubernetes will maintain when you apply the manifest.

#### Events

An event is automatically generated in response to changes with other resources—like nodes, pods, or containers. For example, phases across a pod’s lifecycle transition from pending to running, or statuses like successful or failed may trigger a K8s event. The same goes for re-allocations and scheduling. These events are available until 15 minutes of deployment of the application.

#### Logs

Logs contain the logs of the Pods and Containers deployed which you can use for the process of debugging.


### Create Kubernetes Resource

**Note**: As a pre-requisite, you must have a basic understanding of Kubernetes Cluster, Resources, Kinds.

You can create a Kubernetes resource by passing definition YAML file. You can create more than one resource by separating the resource YAMLs by ‘---’.

An example that shows the required fields and object specifications for a Kubernetes Deployment:

```bash
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

* Click `Create` button on the upper right corner of the `Kubernetes Resource Browser` page.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/create-resource.jpg)

* Provide YAML containing K8s resource configuration and click `Apply`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/create-kubernetes-resource-latest.jpg)

* You will see the details of `Kind`, `Name`, `Status` and `Message` of the created resources.

>Note: A message is displayed only when there is an error in the resource YAML.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/edit-yaml.jpg)


* If required, click `Edit YAML` to edit the YAML or click `Close`.

* A new resource will be created or updated accordingly.





