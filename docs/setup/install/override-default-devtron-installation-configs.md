# Override Default Configurations of Devtron Installation
 
In certain cases, you may want to override default configurations provided by Devtron. For example, for deployments or statefulsets you may want to change the memory or CPU requests or limit or add node affinity or taint tolerance. Say, for ingress, you may want to add annotations or host. Samples are available inside the [manifests/updates](https://github.com/devtron-labs/devtron/tree/main/manifests/updates) directory.
 
To modify a particular object, it looks in namespace `devtroncd` for the corresponding configmap as mentioned in the mapping below:
 
|component| configmap name| purpose|
|-|-|-|
|argocd| argocd-override-cm| GitOps |
|clair|clair-override-cm| container vulnerability db|
|clair| clair-config-override-cm| Clair configuration|
|dashboard| dashboard-override-cm| UI for Devtron|
|gitSensor| git-sensor-override-cm| microservice for Git interaction|
|guard| guard-override-cm| validating webhook to block images with security violations|
|postgresql| postgresql-override-cm| db store of Devtron|
|imageScanner| image-scanner-override-cm| image scanner for vulnerability|
|kubewatch| kubewatch-override-cm| watches changes in ci and cd running in different clusters|
|lens| lens-override-cm| deployment metrics analysis|
|natsOperator| nats-operator-override-cm| operator for nats|
|natsServer| nats-server-override-cm| nats server|
|natsStreaming| nats-streaming-override-cm| nats streaming server|
|notifier| notifier-override-cm| sends notification related to CI and CD |
|devtron| devtron-override-cm| core engine of Devtron|
|devtronIngress| devtron-ingress-override-cm| ingress configuration to expose Devtron|
|workflow| workflow-override-cm| component to run CI workload|
|externalSecret| external-secret-override-cm| manage secret through external stores like vault/AWS secret store|
|grafana| grafana-override-cm| Grafana config for dashboard|
|rollout| rollout-override-cm| manages blue-green and canary deployments|
|minio| minio-override-cm| default store for CI logs and image cache|
|minioStorage| minio-storage-override-cm| db config for minio|
 
Let's take an example to understand how to override specific values. Say, you want to override annotations and host in the ingress, i.e., you want to change devtronIngress, copy the file [devtron-ingress-override.yaml](https://github.com/devtron-labs/devtron/tree/main/manifests/updates/devtron-ingress-override.yaml). This file contains a configmap to modify devtronIngress as mentioned above. Please note the structure of this configmap, data should have the key `override` with a multiline string as a value.
 
`apiVersion`, `kind`, `metadata.name` in the multiline string is used to match the object which needs to be modified. In this particular case it will look for `apiVersion: extensions/v1beta1`, `kind: Ingress` and `metadata.name: devtron-ingress` and will apply changes mentioned inside `update:` as per the example inside the `metadata:` it will add annotations `owner: app1` and inside `spec.rules.http.host` it will add `http://change-me`.
 
In case you want to change multiple objects, for eg in `argocd` you want to change the config of `argocd-dex-server` as well as `argocd-redis` then follow the example in [devtron-argocd-override.yaml](https://github.com/devtron-labs/devtron/tree/main/manifests/updates/devtron-argocd-override.yaml).
 
Once we have made these changes in our local system we need to apply them to a Kubernetes cluster on which Devtron is installed currently using the below command:
 
```bash
kubectl apply -f file-name -n devtroncd
```

Run the following command to make these changes take effect:
 
```bash
kubectl patch -n devtroncd installer installer-devtron --type='json' -p='[{"op": "add", "path": "/spec/reSync", "value": true }]'
```

Our changes would have been propagated to Devtron after 20-30 minutes.
 
## Recommended Resources for Production use

To use Devtron for production deployments, use our recommended production overrides located in [manifests/updates/production](https://github.com/devtron-labs/devtron/tree/main/manifests/updates/production). This configuration should be enough for handling up to 200 microservices.
 
The overall resources required for the recommended production overrides are:
 
|Name| Value|
|-|-|
|cpu| 6|
|memory|13GB|
 
The production overrides can be applied as `pre-devtron installation` as well as `post-devtron installation` in the respective namespace.
 
### Pre-Devtron Installation

If you want to install a new Devtron instance for production-ready deployments, this is the best option for you.
 
Create the namespace and apply the overrides files as stated above:

```bash
kubectl create ns devtroncd
```
 
After files are applied, you are ready to install your Devtron instance with production-ready resources.
 
### Post-Devtron Installation

If you have an existing Devtron instance and want to migrate it for production-ready deployments, this is the right option for you.
 
In the existing namespace, apply the production overrides as we do it above.

```bash
kubectl apply -f prod-configs -n devtroncd
```
