# Upgrading existing devtron to beta

If you want to check the current version of Devtron you are using, please use the following command.

```
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.data}' | grep "^LTAG=" | cut -d"=" -f2-
```

### To upgrade your existing devtron installation to beta, use helm upgrade.

```bash 
$ git clone [https://github.com/devtron-labs/devtron.git](https://github.com/devtron-labs/devtron.git)
$ cd devtron/charts/devtron
$ helm dependency up
$ #modify values in values.yaml
$ helm upgrade devtron . --reuse-values --namespace devtroncd \
-f devtron-bom.yaml
```

> Note: There is no option to upgrade to beta on stack manager UI as of now and you may always see upgrade available for latest stable version using which you'll be moved to latest stable version available.
