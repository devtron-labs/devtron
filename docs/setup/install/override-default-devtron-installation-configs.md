# Override Default Configurations of Devtron Installation

In certain cases you may want to override default configurations provided by Devtron for eg For deployments or statefulsets you want to change memory or cpu requests or limit or may want to node affinity or add taint tolerance or for ingress you may want to add annotations or host. Samples are available inside the [manifests/updates](https://github.com/devtron-labs/devtron/tree/main/manifests/updates) directory. 

In order to modify particular object it looks in namespace `devtroncd` for corresponding configmap as mentioned in mapping below
|-|-|
|component| configmap name| purpose|
|-|-|-|
|argocd| argocd-override-cm| gitops|
|clair|clair-override-cm| container vulnerability db|
|clair| clair-config-override-cm| clair configuration|
|dashboard| dashboard-override-cm| ui of devtron|
|gitSensor| git-sensor-override-cm| microservice for git interaction|
|guard| guard-override-cm| validaing webhook to block images with security violations|
|postgresql| postgresql-override-cm| db store of devtron|
|imageScanner| image-scanner-override-cm| image scanner for vulnerability|
|kubewatch| kubewatch-override-cm| watches changes in ci and cd running in different clusters|
|lens| lens-override-cm| deployment metrics analysis|
|natsOperator| nats-operator-override-cm| operator for nats|
|natsServer| nats-server-override-cm| nats server|
|natsStreaming| nats-streaming-override-cm| nats streaming server|
|notifier| notifier-override-cm| sends notification related to CI and CD|
|devtron| devtron-override-cm| core engine of devtron|
|devtronDexIngress| devtron-dex-ingress-override-cm| ingress configuration to expose devtron|
|workflow| workflow-override-cm| component to run CI workload|
|externalSecret| external-secret-override-cm| manage secret through external stores like vault/AWS secret store|
|grafana| grafana-override-cm| grafana config for dashboard|
|rollout| rollout-override-cm| manages blue-green and canary deployments|
|minio| minio-override-cm| default store for CI logs and image cache|
|minioStorage| minio-storage-override-cm| db config for minio|
|-|-|

Let us take an example to understand how we can override specific values. Assuming you want to override annotations and host in the ingress, that means you want to change devtronDexIngress, please copy file [devtron-ingress-override.yaml](https://github.com/devtron-labs/devtron/tree/main/manifests/updates/devtron-ingress-override.yaml). This file contains a configmap to modify devtronDexIngress as mentioned above. Please note the structure of this configmap, data should have key `override` with multiline string as value. 

`apiVersion`, `kind`, `metadata.name` in the multiline string are used to match the object which needs to be modified. In this particular case it will look for `apiVersion: extensions/v1beta1`, `kind: Ingress` and `metadata.name: devtron-ingress` and will apply changes mentioned inside `update:` as per the example inside the `metadata:` it will add annotations `owner: app1` and inside `spec.rules.http.host` it will add `http://change-me`.

In case you want to change multiple objects, for eg in `argocd` you want to change config of `argocd-dex-server` as well as `argocd-redis` then follow the example in [devtron-argocd-override.yaml](https://github.com/devtron-labs/devtron/tree/main/manifests/updates/devtron-argocd-override.yaml).

Once we have made these changes in our local system we need to apply them to kubernetes cluster on which devtron is installed currently using the below command


```bash
kubectl apply -f file_name -n devtroncd
```

For these changes to come into effect we need to run the following command.

```bash
kubectl patch -n devtroncd installer installer-devtron --type='json' -p='[{"op": "add", "path": "/spec/reSync", "value": true }]'
```

After 20-30 mins our changes would have been propogated to devtron installation.
