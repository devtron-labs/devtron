# Install Devtron on Minikube, Microk8s, K3s, Kind
You can install and try Devtron with CI/CD integration on a high-end Laptop or on a Cloud VM but the laptop/PC may start to respond slow, so it is recommended to uninstall Devtron from your system before shutting it down.

#### System Configurations for Devtron Installation
1. 2 CPUs+ cores
2. 4GB+ of free memory
3. 20GB+ free disk space

## Before you begin
Before we get started and install Devtron, we need to set up the cluster in our servers & install required tools
 * Create cluster using [Minikube](https://minikube.sigs.k8s.io/docs/start/)
 * Create cluster using [Kind tool](https://kind.sigs.k8s.io/docs/user/quick-start/)
 * Create cluster using [K3s](https://rancher.com/docs/k3s/latest/en/installation/)
 * Install [Helm3](https://helm.sh/docs/intro/install/)
 * Install [kubectl](https://kubernetes.io/docs/tasks/tools/)


## Install Devtron on your machine
1. Add Devtron repository
2. Install Devtron 
3. Port-forward the devtron-service to access dashboard

{% tabs %}

{% tab title=" Minikube/kind cluster" %}

 To install devtron on ``Minikube/kind`` Cluster use the Following commands
```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort 

```
{% endtab %}

{% tab title="k3s Cluster" %}
To install devtron on ``k3s`` Cluster use the Following commands
```bash
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml

helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort

```
{% endtab %}

{% endtabs %}
 
### Devtron dashboard

To access dashboard when using ``Minikube`` as Cluster use this command, dashboard will automatically open on default browser.

```bash
minikube service devtron-service --namespace devtroncd
```

To access dashboard when using ``Kind/k3s`` as Cluster, use this command to port forward the devtron service to port 8000  
```bash
kubectl -ndevtroncd port-forward service/devtron-service 8000:80
```
Dashboard [http://127.0.0.1:8000](http://127.0.0.1:8000).

### Devtron Admin credentials

For admin login, use the username:`admin`, and run the following command to get the admin password:

```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```

## Install Devtron on Cloud VM (AWS ec2, Azure VM, GCP VM)
It is preferd to use Cloud VM with 2vCPU+, 4GB+ free Memory, 20GB+ Storage, Compute Optimized VM type & Ubuntu Flavoured OS.
 1. Create Microk8s Cluster
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


2. Install devtron
```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort 

```
3. Ensure that the port on which the devtron-service runs is open in the VM's security group or network Security group.

Commad to get the devtron-service Port number
```bash
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.spec.ports[0].nodePort}'
```
