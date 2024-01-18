# Install Devtron on Minikube, Microk8s, K3s, Kind, Cloud VMs

You can install and try Devtron on a high-end machine or a Cloud VM. If you install it on a laptop/PC, it may start to respond slowly, so it is recommended to uninstall Devtron from your system before shutting it down.

## Prerequisites

1. 2 vCPUs
2. 4GB+ of free memory
3. 20GB+ free disk space

Before you get started, you must set up a cluster in your server and finish the following actions:

 * Create a cluster using [Minikube](https://minikube.sigs.k8s.io/docs/start/) or [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) or [K3s](https://rancher.com/docs/k3s/latest/en/installation/).
 * Install [Helm3](https://helm.sh/docs/intro/install/).
 * Install [kubectl](https://kubernetes.io/docs/tasks/tools/).

---

## Tutorial 

{% embed url="https://www.youtube.com/watch?v=rKUymNJqcjA" caption="Installing Devtron on Minikube" %}

---

## For Minikube, Microk8s, K3s, Kind

{% tabs %}

{% tab title=" Minikube/Kind Cluster" %}

 To install devtron on ``Minikube/kind`` cluster, run the following command:

```bash
helm repo add devtron https://helm.devtron.ai

helm repo update devtron

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort --set installer.arch=multi-arch

```
{% endtab %}

{% tab title="k3s Cluster" %}
To install devtron on ``k3s`` cluster, run the following command:

```bash
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml

helm repo add devtron https://helm.devtron.ai

helm repo update devtron

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort

```
{% endtab %}

{% endtabs %}
 
### Access Devtron Dashboard

To access Devtron dashboard when using ``Minikube`` as cluster, run the following command:

```bash
minikube service devtron-service --namespace devtroncd
```

To access Devtron dashboard when using ``Kind/k3s`` as cluster, run the following command to port forward the devtron service to port 8000:

```bash
kubectl -n devtroncd port-forward service/devtron-service 8000:80
```

**Dashboard**: [http://127.0.0.1:8000](http://127.0.0.1:8000).

### Get Admin Credentials

When you install Devtron for the first time, it creates a default admin user and password (with unrestricted access to Devtron). You can use that credentials to log in as an administrator. 

After the initial login, we recommend you set up any SSO service like Google, GitHub, etc., and then add other users (including yourself). Subsequently, all the users can use the same SSO (let's say, GitHub) to log in to Devtron's dashboard.

The section below will help you understand the process of getting the administrator credentials.

#### For Devtron version v0.6.0 and higher

**Username**: `admin` <br>
**Password**: Run the following command to get the admin password:

```bash
kubectl -n devtroncd get secret devtron-secret \
-o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
```


<details>
<summary>For Devtron version less than v0.6.0</summary>

**Username**: `admin` <br>
**Password**: Run the following command to get the admin password:

```bash
kubectl -n devtroncd get secret devtron-secret \
-o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```
</details>

---

## For Cloud VM (AWS EC2, Azure VM, GCP VM)

It is recommended to use Cloud VM with 2vCPU+, 4GB+ free memory, 20GB+ storage, Compute Optimized VM type & Ubuntu Flavoured OS.

### Create Microk8s Cluster

```bash
sudo snap install microk8s --classic --channel=1.22
sudo usermod -a -G microk8s $USER
sudo chown -f -R $USER ~/.kube
newgrp microk8s
microk8s enable dns storage helm3
echo "alias kubectl='microk8s kubectl '" >> .bashrc
echo "alias helm='microk8s helm3 '" >> .bashrc
source .bashrc
```

### Install Devtron

```bash
helm repo add devtron https://helm.devtron.ai

helm repo update devtron

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort 

```
### Get devtron-service Port Number

```bash
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.spec.ports[0].nodePort}'
```

Make sure that the port on which the devtron-service runs remain open in the VM's security group or network security group.

{% hint style="info" %}
If you want to uninstall Devtron or clean Devtron helm installer, refer our [uninstall Devtron](./uninstall-devtron.md).
{% endhint %}

If you have questions, please let us know on our discord channel. [![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)
