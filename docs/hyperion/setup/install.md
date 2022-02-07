# Install Hyperion using Helm3

To install Helm3, please check [Installing Helm3](https://helm.sh/docs/intro/install/)

{% tabs %}
{% tab title="Install with default configurations" %}
```bash
helm repo add devtron https://helm.devtron.ai
helm install devtron devtron/devtron-operator --create-namespace --namespace devtroncd --set installer.mode=hyperion
```
{% endtab %}
{% endtabs %}

For those countries/users where Github is blocked , you can download the [Hyperion Helm chart](https://s3-ap-southeast-1.amazonaws.com/devtron.ai/devtron-operator-latest.tgz)


```bash
wget https://s3-ap-southeast-1.amazonaws.com/devtron.ai/devtron-operator-latest.tgz
helm install devtron devtron-operator-latest.tgz --create-namespace --namespace devtroncd --set installer.mode=hyperion
```

[//]: # (If you are planning to use Hyperion for `production deployments`, please refer to our recommended overrides for [Devtron Installation]&#40;override-default-devtron-installation-configs.md&#41;.)

[//]: # (## Installation status)

[//]: # ()
[//]: # (Run following command)

[//]: # ()
[//]: # (```bash)

[//]: # (kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}')

[//]: # (```)

## Access Hyperion dashboard

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

### Hyperion Admin credentials

For admin login use username:`admin` and for password run the following command.

```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```

### Cleaning Hyperion Helm3

Please make sure that you do not have anything inside namespaces devtroncd, devtron-cd devtron-ci and devtron-demo as the below steps will clean everything inside these namespaces
```
helm uninstall devtron --namespace devtroncd
kubectl delete -n devtroncd -f https://raw.githubusercontent.com/devtron-labs/charts/main/charts/devtron/crds/crd-devtron.yaml
kubectl delete ns devtroncd
```
