# Upgrading Devtron 0.3.x to 0.3.x

## Follow the required step to update the Devtron version

Set `reSync: true` in the installer object, this will initiate upgrade of the entire Devtron stack, you can use the following command to do this.

```bash
kubectl patch -n devtroncd installer installer-devtron --type='json' -p='[{"op": "add", "path": "/spec/reSync", "value": true }]'
```

## Follow the required step if you want to upgrade using HELM

```bash
helm repo update
helm upgrade devtron devtron/devtron-operator --namespace devtroncd
```
