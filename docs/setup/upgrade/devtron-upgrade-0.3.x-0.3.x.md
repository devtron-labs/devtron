# Upgrading Devtron 0.3.x to 0.3.x

If you want to check the current version of Devtron you are using, please use the following command.

```
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.data}' | grep "^LTAG=" | cut -d"=" -f2-
```

## Follow the required step to upgrade the Devtron version

```bash
helm repo update
helm upgrade devtron devtron/devtron-operator --namespace devtroncd --set installer.release=v0.3.x
```

You can find the latest releases from Devtron Github Repository:

https://github.com/devtron-labs/devtron/releases
