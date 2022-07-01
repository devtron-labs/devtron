# Install Devtron on Local Machine/VM

## Before you begin
Before we get started and install Devtron, we need to set up the cluster in our servers.
 * Create cluster using [Minikube](https://minikube.sigs.k8s.io/docs/start/)
 * Create cluster using [Kind tool](https://kind.sigs.k8s.io/docs/user/quick-start/)
 * Create cluster using [K3s](https://rancher.com/docs/k3s/latest/en/installation/)
 * Install [Helm3](https://helm.sh/docs/intro/install/)
 * Install [kubectl](https://kubernetes.io/docs/tasks/tools/)
#### System Configurations for Devtron Installation
1. 2 CPUs+ cores
2. 4GB+ of free memory
3. 20GB+ free disk space

## Installing Devtron on Minikube/Kind Cluster
1. Add Devtron repository
2. Install Devtron 
3. Port-forward the devtron-service to access dashboard

{% tabs %}
{% tab title="Devtron on Minikube/kind cluster" %}

```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort

```
{% endtab %}


{% tab title="Devtron on k3s Cluster" %}

```bash
curl -sfL https://get.k3s.io | sh -s - --write-kubeconfig-mode 644

kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml

helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort

```

{% endtab %}
{% endtab %}
 

### Devtron Admin credentials

For admin login, use the username:`admin`, and run the following command to get the admin password:

```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```

### Devtron dashboard

To access dashboard when using ``Minikube`` as Cluster use this command, dashboard will automatically open on default browser

```bash
minikube service devtron-service --namespace devtroncd
```

To access dashboard when using ``Kind/k3s`` as Cluster use this command to port forward the devtron service to port 8000  
```bash
kubectl -ndevtroncd port-forward service/devtron-service 8000:80
```
Dashboard [http://127.0.0.1:8000](http://127.0.0.1:8000).