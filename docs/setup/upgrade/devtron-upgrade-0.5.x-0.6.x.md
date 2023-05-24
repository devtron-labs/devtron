# Upgrading Devtron 0.5.x to 0.6.x

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
export RELEASE_NAME=devtron
```

### 3. Run the following script to upgrade

```bash
wget https://raw.githubusercontent.com/devtron-labs/utilities/main/scripts/shell/upgrade-devtron-v6.sh
```

```bash
sh upgrade-devtron-v6.sh
```
Please ignore any errors you encounter while running the upgrade script