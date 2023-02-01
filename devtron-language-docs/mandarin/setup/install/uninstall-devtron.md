#卸载Devtron

要卸载Devtron，请运行以下命令：

此命令将删除与Devtron相关的所有命名空间（`devtroncd`、`devtron-cd`、`devtron-ci` 等）。

\```bash

helm uninstall devtron --namespace devtroncd

kubectl delete -n devtroncd -f https://raw.githubusercontent.com/devtron-labs/charts/main/charts/devtron/crds/crd-devtron.yaml

kubectl delete -n argo -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/yamls/workflow.yaml

kubectl delete ns devtroncd devtron-cd devtron-ci devtron-demo argo

\```


\*\*注意\*\*：如果您有任何疑问，请通过我们的Discord频道告诉我们。 [![加入 Discord] (https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)

