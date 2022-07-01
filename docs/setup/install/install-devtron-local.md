# Install Devtron on Local Machine/VM

## Before you begin

Install [Helm3](https://helm.sh/docs/intro/install/).

## Installing Devtron on Minikube/Kind Cluster
1. Add Devtron repository
2. Install Devtron 
3. Port-forward the devtron-service to access dashboard

{% tabs %}
{% tab title="Devtron on Minikube cluster" %}

```bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort

```
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

To access dashboard when using ``Kind`` as Cluster use this command to port forward the devtron service to port 8000  
```bash
kubectl -ndevtroncd port-forward service/devtron-service 8000:80
```
Dashboard [http://127.0.0.1:8000](http://127.0.0.1:8000).