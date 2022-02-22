# Upgrading Devtron 0.3.x to 0.3.x

If you want to check the current version of Devtron you are using, please use the following command.

```
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.data}' | grep "^LTAG=" | cut -d"=" -f2-
```

## Follow the required step to upgrade the Devtron version

Set `reSync: true` in the installer object, this will initiate upgrade of the entire Devtron stack, you can use the following command to do this.

```bash
kubectl patch -n devtroncd installer installer-devtron --type='json' -p='[{"op": "add", "path": "/spec/reSync", "value": true }]'
```

## Follow the required step if you want to upgrade using HELM

```bash
helm repo update
helm upgrade devtron devtron/devtron-operator --namespace devtroncd
```
