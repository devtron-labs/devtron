### Install Devtron using kubectl

If you don't want to install helm and just want to use `kubectl` to install `devtron platform`, then please follow the steps mentioned below:

```bash
wget https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/install/devtron-operator-configs.yaml
```
Use a preferred editor to edit the values in install/devtron-operator-configs.yaml
```bash
vim devtron-operator-configs.yaml
```
Edit the `devtron-operator-configs.yaml` to configure your Devtron installation. For more details about it, see [configuration](#configuration)
Once your configurations are ready, continue with following steps
```bash
kubectl create ns devtroncd
kubectl -n devtroncd apply -f https://raw.githubusercontent.com/devtron-labs/charts/main/charts/devtron/crds/crd-devtron.yaml
kubectl apply -n devtroncd -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/install/install.yaml
```
Edit devtron-operator-configs.yaml and input the values
```bash
kubectl apply -n devtroncd -f devtron-operator-configs.yaml
kubectl apply -n devtroncd -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/install/devtron-installer.yaml
```

## Installation status

Run following command

```bash
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}'
```

The install commands initiates Devtron-operator which spins up all the Devtron micro-services one by one in about 30 mins. You can use the above command to check the status of the installation if the installation is still in progress, it will print `Downloaded`. When the installation is complete, it prints `Applied`.

## Access Devtron dashboard

If you did not provide a **BASE\_URL** during install or have used the default installation, Devtron creates a loadbalancer for you on its own. Use the following command to get the dashboard url.

```text
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
```

You will get result something like below

```text
[test2@server ~]$ kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
[map[hostname:aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com]]
```

The hostname mentioned here \( aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com \) is the Loadbalancer URL where you can access the Devtron dashboard.

**PS:** You can also do a CNAME entry corresponding to your domain/subdomain to point to this Loadbalancer URL to access it at a custom domain.

| Host | Type | Points to |
| ---: | :--- | :--- |
| devtron.yourdomain.com | CNAME | aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com |

### Devtron Admin credentials

For admin login use username:`admin` and for password run the following command.

```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```


### Cleaning Devtron installer
Please make sure that you do not have anything inside namespace devtroncd as the below steps will clean everything inside namespace devtroncd
```bash
kubectl delete -n devtroncd -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/install/devtron-installer.yaml
kubectl delete -n devtroncd -f https://raw.githubusercontent.com/devtron-labs/charts/main/charts/devtron/templates/install.yaml
kubectl delete -n devtroncd -f https://raw.githubusercontent.com/devtron-labs/charts/main/charts/devtron/crds/crd-devtron.yaml
kubectl delete ns devtroncd
```
