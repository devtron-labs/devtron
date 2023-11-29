# Install Devtron

In this section, we describe on how you can install Helm Dashboard by Devtron without any integrations. Integrations can be added later using [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations).

If you want to install Devtron on Minikube, Microk8s, K3s, Kind, refer this [section](./Install-devtron-on-Minikube-Microk8s-K3s-Kind.md).

## Before you begin

Install [Helm](https://helm.sh/docs/intro/install/) if you have not installed it.

## Add Helm Repo

```bash
helm repo add devtron https://helm.devtron.ai
```

## Install Helm Dashboard by Devtron

**Note**: This installation command will not install CI/CD integration. For CI/CD, refer [install Devtron with CI/CD](https://docs.devtron.ai/install/install-devtron-with-cicd) section.

Run the following command to install Helm Dashboard by Devtron:

```bash
helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd
```


## Install Multi-Architecture Nodes (ARM and AMD)

To install Devtron on clusters with the multi-architecture nodes (ARM and AMD), append the Devtron installation command with `--set installer.arch=multi-arch`.



[//]: # (If you are planning to use Hyperion for `production deployments`, please refer to our recommended overrides for [Devtron Installation]&#40;override-default-devtron-installation-configs.md&#41;.)

[//]: # (## Installation status)

[//]: # ()
[//]: # (Run following command)

[//]: # ()
[//]: # (```bash)

[//]: # (kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}')

[//]: # (```)

## Devtron Dashboard

### For MiniKube/K3s/Kind/Microk8s/On-prem

Run the below Command:

```bash
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.spec.ports[0].nodePort}'
```
Access your Devtron Dashboard with  ```machineIP:nodePort```

For more information, refer [Install Devtron on Minikube, Microk8s, K3s, Kind](https://docs.devtron.ai/install/install-devtron-on-minikube-microk8s-k3s-kind) 

Run the following command to get the dashboard URL:

```text
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
```

You will get the result something as shown below:

```text
[test2@server ~]$ kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
[map[hostname:aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com]]
```

The hostname `aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com` as mentioned above is the Loadbalancer URL where you can access the Devtron dashboard.

> You can also do a CNAME entry corresponding to your domain/subdomain to point to this Loadbalancer URL to access it at a custom domain.

| Host | Type | Points to |
| :--- | :--- | :--- |
| devtron.yourdomain.com | CNAME | aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com |


## Devtron Admin credentials

When you install Devtron for the first time, it creates a default admin user and password (with unrestricted access to Devtron). You can use that credentials to log in as an administrator. 

After the initial login, we recommend you set up any SSO service like Google, GitHub, etc., and then add other users (including yourself). Subsequently, all the users can use the same SSO (let's say, GitHub) to log in to Devtron's dashboard.

The section below will help you understand the process of getting the administrator credentials.

### For Devtron version v0.6.0 and higher

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


**Note**: If you want to uninstall Devtron or clean Devtron helm installer, refer our [uninstall Devtron](https://docs.devtron.ai/install/uninstall-devtron).


## Upgrade

To use the CI/CD capabilities with Devtron, you can Install the [Devtron with CI/CD](https://docs.devtron.ai/install/install-devtron-with-cicd) or [Devtron with CI/CD along with GitOps (Argo CD)](https://docs.devtron.ai/install/install-devtron-with-cicd-with-gitops).
