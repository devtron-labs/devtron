# Upgrading Devtron 0.3.x to 0.3.x

If you want to check the current version of Devtron you are using, please use the following command.

```
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.data}' | grep "^LTAG=" | cut -d"=" -f2-
```

## Follow the below mentioned steps to upgrade the Devtron version

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
