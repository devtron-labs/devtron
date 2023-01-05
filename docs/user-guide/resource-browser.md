# Resource Browser

`Resource Browser` lists all of the resources running in cluster in your current project. You can use it to view, inspect, manage, and delete resources in your clusters. You can also create resources from the `Resource Browser`.

Resource Browser are helpful for troubleshooting issues. It supports multi-cluster.

## Kubernetes Resources 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/kubernetes-resource-browser-latest.jpg)

The following resources are grouped in the following categories:

* **Workloads** displays workloads (Cronjob, Deployment, StatefulSet, DaemonSet, Job, and Pod resources) deployed to clusters in your current project. Includes each workload's name, status, type, number of running and total desired Pods, namespace, and cluster.

* **Config & Storage** provide both long-term and temporary [storage](https://kubernetes.io/docs/concepts/storage/) to Pods in your cluster and [configuration](https://kubernetes.io/docs/concepts/configuration/) of pods.

* **Networking** displays your project's [Enpoints](https://kubernetes.io/docs/concepts/services-networking/endpoint-slices/), [NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/) and [Service](https://kubernetes.io/docs/concepts/services-networking/service/) resources. Displays each resource's name, namespace, type, Cluster-IP, External-IP, Port(s), Age.

* **RBAC** stands for Role-based access control which provides the authorization strategy for regulating access to cluster or resources based on the roles of individual users within your organization.

* **Custom Resource** allows you to create your own API resources and define your own kind just like Pod, Deployment, ReplicaSet, etc. 


The following resources are grouped as uncategoried:

** **Events** displays all the reports of an event in a cluster.

** **Namespaces** displays the current list of namespaces in a cluster.

### Search and Filter Resources

You can search and filter resources by specific resource Kinds. You can also preview `Manifest`, `Events`, `Logs`, access `Terminal` by selecting ellipsis on the specific resource or `Delete` a specific resource.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/select-resource.jpg)

#### Manifest

The Manifest shows the critical information such as container-image, restartCount, state, phase, podIP, startTime etc. and status of the pods which are deployed.

#### Events

Events display you the events that took place during the deployment of an application. These events are available until 15 minutes of deployment of the application.

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





