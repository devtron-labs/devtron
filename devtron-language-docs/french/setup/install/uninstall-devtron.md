\# Désinstaller Devtron

Pour désinstaller Devtron, exécutez la commande suivante :

Cette commande supprimera tous les namespaces liés à Devtron (`devtroncd`, `devtron-cd`, `devtron-ci` etc.).

\```bash

helm uninstall devtron --namespace devtroncd

kubectl delete -n devtroncd -f https://raw.githubusercontent.com/devtron-labs/charts/main/charts/devtron/crds/crd-devtron.yaml

kubectl delete -n argo -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/yamls/workflow.yaml

kubectl delete ns devtroncd devtron-cd devtron-ci devtron-demo argo

\```


\*\*Note\*\* : Si vous avez des questions, veuillez nous en faire part sur notre canal discord. [![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)

