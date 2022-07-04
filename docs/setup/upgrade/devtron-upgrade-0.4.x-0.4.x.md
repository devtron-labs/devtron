# Upgrading Devtron 0.4.x to 0.4.x

If you want to check the current version of Devtron you are using, please use the following command.

```
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.data}' | grep "^LTAG=" | cut -d"=" -f2-
```

## Follow the below mentioned steps to upgrade the Devtron version using Helm

### 1. Check the devtron release name

```bash
helm list --namespace devtroncd
```

### 2. Set release name in the variable
```bash
RELEASE_NAME=devtron
```

### 3. Fetch the latest Devtron helm chart

```bash
helm repo update
```

### 4. Annotate and Label Devtron resources

```bash
kubectl -n devtroncd label role --all "app.kubernetes.io/managed-by=Helm"
kubectl -n devtroncd annotate role --all "meta.helm.sh/release-name=$RELEASE_NAME" "meta.helm.sh/release-namespace=devtroncd"
kubectl -n devtroncd label rolebinding --all "app.kubernetes.io/managed-by=Helm"
kubectl -n devtroncd annotate rolebinding --all "meta.helm.sh/release-name=$RELEASE_NAME" "meta.helm.sh/release-namespace=devtroncd"
```

### 5. Upgrade Devtron 

5.1 Upgrade Devtron to latest version

```bash
helm upgrade devtron devtron/devtron-operator --namespace devtroncd \
-f https://raw.githubusercontent.com/devtron-labs/devtron/main/charts/devtron/devtron-bom.yaml \
--set installer.modules={cicd} --reuse-values
```
OR

5.2 Upgrade Devtron to a custom version. You can find the latest releases from Devtron on Github https://github.com/devtron-labs/devtron/releases

```bash
DEVTRON_TARGET_VERSION=v0.4.x

helm upgrade devtron devtron/devtron-operator --namespace devtroncd \
-f https://raw.githubusercontent.com/devtron-labs/devtron/$DEVTRON_TARGET_VERSION/charts/devtron/devtron-bom.yaml \
--set installer.modules={cicd} --reuse-values
```
