\# Devtron deinstallieren

Führen Sie den folgenden Befehl aus, um Devtron zu deinstallieren:

Dieser Befehl entfernt alle Namespaces, die sich auf Devtron beziehen (`devtroncd`, `devtron-cd`, `devtron-ci` usw.).

\```bash

helm uninstall devtron --namespace devtroncd

kubectl delete -n devtroncd -f https://raw.githubusercontent.com/devtron-labs/charts/main/charts/devtron/crds/crd-devtron.yaml

kubectl delete -n argo -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/yamls/workflow.yaml

kubectl delete ns devtroncd devtron-cd devtron-ci devtron-demo argo

\```


\*\*Hinweis\*\*: Wenn Sie Fragen haben, melden Sie sich bitte in unserem Discord-Channel. [![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)

