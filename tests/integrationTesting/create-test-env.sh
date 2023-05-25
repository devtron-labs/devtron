# shellcheck disable=SC2155
apk update && apk add wget && apk add curl && apk add vim  && apk add bash && apk add git && apk add yq
wget -q -O - https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
k3d cluster create it-cluster
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
wget https://raw.githubusercontent.com/devtron-labs/devtron/main/tests/integrationTesting/postgresql-secret.yaml -O postgresql-secret.yaml
kubectl apply -f postgresql-secret.yaml
wget https://raw.githubusercontent.com/devtron-labs/devtron/main/tests/integrationTesting/postgresql.yaml -O postgresql.yaml
kubectl apply -f postgresql.yaml
wget https://raw.githubusercontent.com/devtron-labs/devtron/main/tests/integrationTesting/migrator.yaml -O migrator.yaml
yq '(select(.metadata.name == "postgresql-migrate-devtron") | .spec.template.spec.containers[0].env[0].value) = env(TEST_BRANCH)' migrator.yaml -i
yq '(select(.metadata.name == "postgresql-migrate-devtron") | .spec.template.spec.containers[0].env[9].value) = env(LATEST_HASH)' migrator.yaml -i
kubectl apply -f migrator.yaml
# shellcheck disable=SC2046
while [ ! $(kubectl -n devtroncd get $(kubectl -n devtroncd get job -l job=postgresql-migrate-devtron -o name) -o jsonpath="{.status.succeeded}")  ]; do sleep 10; done
# shellcheck disable=SC2046
while [ ! $(kubectl -n devtroncd get $(kubectl -n devtroncd get job -l job=postgresql-migrate-casbin -o name) -o jsonpath="{.status.succeeded}")  ]; do sleep 10; done
# shellcheck disable=SC2046
while [ ! $(kubectl -n devtroncd get $(kubectl -n devtroncd get job -l job=postgresql-migrate-lens -o name) -o jsonpath="{.status.succeeded}")  ]; do sleep 10; done
# shellcheck disable=SC2046
while [ ! $(kubectl -n devtroncd get $(kubectl -n devtroncd get job -l job=postgresql-migrate-gitsensor -o name) -o jsonpath="{.status.succeeded}")  ]; do sleep 10; done
exit #to get out of container