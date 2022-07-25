# Upgrading Devtron 0.4.x to 0.5.x

>🔥 IF YOU ARE USING rawYaml SECTION IN DEPLOYMENT TEMPLATE, THIS RELEASE CAN INTRODUCE BREAKING CHANGES, WE RECOMMEND YOU TO UPDATE THE CHART VERSION OF YOUR APP USING rawYaml TO v4.13.0 TO MAKE rawYaml SECTION COMPATIBLE TO NEW ARGOCD VERSION v2.4.0
> 
> OR 
> 
> APPLY THE AUTOMATED PATCH TO FIX THE ISSUE AND THEN PROCEED TO UPGRADE YOUR DEVTRON STACK TO v0.5.0
> 
> CONTACT DEVTRON TEAM ON DISCORD TO DISCUSS YOUR USECASE.



If you want to check the current version of Devtron you are using, please use the following command.

```
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.data}' | grep "^LTAG=" | cut -d"=" -f2-
```

## Follow the below mentioned steps to upgrade the Devtron version using Helm

### 1. Run the Devtron v0.5.0 pre-upgrade patch job
```
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

### 5. Annotate and Label Devtron resources

```bash
kubectl -n devtroncd label role --all "app.kubernetes.io/managed-by=Helm"
kubectl -n devtroncd annotate role --all "meta.helm.sh/release-name=$RELEASE_NAME" "meta.helm.sh/release-namespace=devtroncd"
kubectl -n devtroncd label rolebinding --all "app.kubernetes.io/managed-by=Helm"
kubectl -n devtroncd annotate rolebinding --all "meta.helm.sh/release-name=$RELEASE_NAME" "meta.helm.sh/release-namespace=devtroncd"
```

### 6. Upgrade Devtron 

6.1 Upgrade Devtron to latest version

```bash
helm upgrade devtron devtron/devtron-operator --namespace devtroncd \
-f https://raw.githubusercontent.com/devtron-labs/devtron/main/charts/devtron/devtron-bom.yaml \
--set installer.modules={cicd} --reuse-values
```
OR

6.2 Upgrade Devtron to a custom version. You can find the latest releases from Devtron on Github https://github.com/devtron-labs/devtron/releases

```bash
DEVTRON_TARGET_VERSION=v0.4.x

helm upgrade devtron devtron/devtron-operator --namespace devtroncd \
-f https://raw.githubusercontent.com/devtron-labs/devtron/$DEVTRON_TARGET_VERSION/charts/devtron/devtron-bom.yaml \
--set installer.modules={cicd} --reuse-values
```
