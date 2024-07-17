# Upgrading Devtron 0.6.x to 0.7.x

To check the current version of your Devtron setup, use the following command

```bash
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.data}' | grep "^LTAG=" | cut -d"=" -f2-
```

Proceed with the following steps only if the version is `0.6.x`

---

## Prerequisites

1. Set the release name

```bash
export RELEASE_NAME=devtron
```

2. Label and annotate the service accounts in the `devtron-ci` namespace

```bash
kubectl -n devtron-ci label sa --all "app.kubernetes.io/managed-by=Helm" --overwrite
kubectl -n devtron-ci annotate sa --all "meta.helm.sh/release-name=$RELEASE_NAME" "meta.helm.sh/release-namespace=devtroncd" --overwrite
```

3. Now, label and annotate the service accounts in the `devtron-cd` namespace

```
kubectl -n devtron-cd label sa --all "app.kubernetes.io/managed-by=Helm" --overwrite
kubectl -n devtron-cd annotate sa --all "meta.helm.sh/release-name=$RELEASE_NAME" "meta.helm.sh/release-namespace=devtroncd" --overwrite
```

---

## Upgrade Commands

1. Update the Helm repository

```bash
helm repo update
```

2. Run the upgrade command for Devtron

```bash
helm upgrade devtron devtron/devtron-operator -n devtroncd --reuse-values -f https://raw.githubusercontent.com/devtron-labs/devtron/main/charts/devtron/devtron-bom.yaml
```

---

## Expected Command Output

![Command Output](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/kubernetes-resource-browser/command-output.jpg)
