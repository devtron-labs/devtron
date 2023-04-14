# Upgrading Devtron 0.4.x to 0.5.x

If you want to check the current version of Devtron you are using, please use the following command.

```
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.data}' | grep "^LTAG=" | cut -d"=" -f2-
```

## Follow the below mentioned steps to upgrade the Devtron version using Helm


### 1. Apply Prerequisites Patch Job

If you are using rawYaml in deployment template, this update can introduce breaking changes. We recommend you to update the `Chart Version`
of your app to `v4.13.0` to make rawYaml section compatible to new argocd version `v2.4.0`.

Or

We have released a argocd-v2.4.0 patch job to fix the compatibilities issues. Please apply this job in your cluster and wait for completion
and then only upgrade to `Devtron v0.5.x`.

```bash
kubectl apply -f https://raw.githubusercontent.com/devtron-labs/utilities/main/scripts/jobs/argocd-2.4.0-prerequisites-patch-job.yaml
```

### 2. Check the devtron release name

```bash
helm list --namespace devtroncd
```

### 3. Set release name in the variable
```bash
RELEASE_NAME=devtron
```

### 4. Fetch the latest Devtron helm chart

```bash
helm repo update
```


### 5. Upgrade Devtron 

**`5.1` Upgrade Devtron to latest version**

```bash
helm upgrade devtron devtron/devtron-operator --namespace devtroncd \
-f https://raw.githubusercontent.com/devtron-labs/devtron/main/charts/devtron/devtron-bom.yaml \
--set installer.modules={cicd} --reuse-values
```
OR

**`5.2` Upgrade Devtron to a custom version**

 You can find the latest releases from Devtron on Github https://github.com/devtron-labs/devtron/releases

```bash
DEVTRON_TARGET_VERSION=v0.5.x

helm upgrade devtron devtron/devtron-operator --namespace devtroncd \
-f https://raw.githubusercontent.com/devtron-labs/devtron/$DEVTRON_TARGET_VERSION/charts/devtron/devtron-bom.yaml \
--set installer.modules={cicd} --reuse-values
```