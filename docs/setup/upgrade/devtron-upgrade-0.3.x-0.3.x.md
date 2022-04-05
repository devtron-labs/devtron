# Upgrading Devtron 0.3.x to 0.3.x

If you want to check the current version of Devtron you are using, please use the following command.

```
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.data}' | grep "^LTAG=" | cut -d"=" -f2-
```

## Follow the below mentioned steps to upgrade the Devtron version using Helm

1. Fetch the latest Devtron helm chart

```bash
helm repo update
```

2. Input the target Devtron version that you want to upgrade to. You can find the latest releases from Devtron on Github https://github.com/devtron-labs/devtron/releases

```bash
DEVTRON_TARGET_VERSION=v0.3.x
```

3. Upgrade Devtron

```bash
helm upgrade devtron devtron/devtron-operator --namespace devtroncd --set installer.release=$DEVTRON_TARGET_VERSION
```


## Follow the below mentioned steps to upgrade the Devtron version using Kubectl

1. Input the target Devtron version that you want to upgrade to. You can find the latest releases from Devtron on Github https://github.com/devtron-labs/devtron/releases

```bash
DEVTRON_TARGET_VERSION=v0.3.x
```

2. Patch Devtron Installer

```bash
kubectl patch -n devtroncd installer installer-devtron --type='json' -p='[{"op": "add", "path": "/spec/reSync", "value": true },{"op": "replace", "path": "/spec/url", "value": "https://raw.githubusercontent.com/devtron-labs/devtron/$DEVTRON_TARGET_VERSION/manifests/installation-script"}]'
```