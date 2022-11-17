# Install Devtron on GKE


Devtron can be installed on any Kubernetes cluster. This cluster can use upstream Kubernetes, or it can be a managed Kubernetes cluster from a cloud provider such as [GKE](https://cloud.google.com/kubernetes-engine/).

In this section, we will walk you through the steps of installing Devtron on [GKE](https://cloud.google.com/kubernetes-engine/). 

## Create a GKE Cluster

A standard Kubernetes cluster in GKE is a prerequisite for installing Devtron. 


* Create a GKE cluster

```bash
create-cluster-gke
```

* In **Cluster basics**, select a Master version. The static version 1.15.12-gke.2 is used here as an example.

```bash
select-master-version
```

* In **default-pool** under **Node Pools**, define 3 nodes in this cluster.

```bash
node-number
```

* Go to **Nodes**, select the image type and set the Machine Configuration as below. When you finish, click **Create**.

```bash
machine-config
```

* When the GKE cluster is ready, you can connect to the cluster with Cloud Shell.

```bash
cloud-shell-gke
```



## Install Devtron with CI/CD on GKE

* Install Devtron with CI/CD with the following command:

```bash
helm repo add devtron https://helm.devtron.ai
```
```bash
helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd}
```

* To check the status of pods, run the following command:

```bash
kubectl get po -n devtroncd 
```

* Or, to track the progress of Devtron microservices installation, run the following command:

```bash
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}'
```
The command executes with one of the following output messages, indicating the status of the installation:

| Status | Description |
| :--- | :--- |
| `Downloaded` | The installer has downloaded all the manifests, and the installation is in progress. |
| `Applied` | The installer has successfully applied all the manifests, and the installation is complete. |

Once the status is `Downloaded`, you can access the Devtron dashoboard URL.

* To get the Devtron dashboard URL, run the following command:

 ```bash
 kubectl get svc -n devtroncd devtron-service \
-o jsonpath='{.status.loadBalancer.ingress}'
```

* You will get an output similar to the example as shown below:

```bash
[test2@server ~]$ kubectl get svc -n devtroncd devtron-service \
-o jsonpath='{.status.loadBalancer.ingress}'
[map[hostname:aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com]]
```
You can access the Devtron dasboard URL using the hostname `aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com` as shown in the above example.


* To get the password for the default admin user, run the following command:
```bash
 kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
 ```