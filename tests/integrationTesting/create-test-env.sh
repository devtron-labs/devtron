# shellcheck disable=SC2155
export TEST_BRANCH=$(echo $TEST_BRANCH | awk -F '/' '{print $NF}')
apk update && apk add wget && apk add curl && apk add vim  && apk add bash && apk add git && apk add yq && apk add gcc && apk add musl-dev && apk add make
wget -q -O - https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
k3d cluster create it-cluster
export MACHINE_OS='linux'
#export MACHINE_OS='darwin'
export MACHINE_ARCH='amd64'
#export MACHINE_ARCH='arm64'
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/$MACHINE_OS/$MACHINE_ARCH/kubectl"
install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
kubectl create ns devtroncd
kubectl create ns devtron-cd
kubectl create ns devtron-ci
#wget https://raw.githubusercontent.com/devtron-labs/devtron/main/tests/integrationTesting/postgresql-secret.yaml -O postgresql-secret.yaml
kubectl -n devtroncd apply -f postgresql-secret.yaml
#wget https://raw.githubusercontent.com/devtron-labs/devtron/main/tests/integrationTesting/postgresql.yaml -O postgresql.yaml
kubectl -ndevtroncd apply -f postgresql.yaml
#wget https://raw.githubusercontent.com/devtron-labs/devtron/main/tests/integrationTesting/migrator.yaml -O migrator.yaml
yq '(select(.metadata.name == "postgresql-migrate-devtron") | .spec.template.spec.containers[0].env[0].value) = env(TEST_BRANCH)' migrator.yaml -i
yq '(select(.metadata.name == "postgresql-migrate-devtron") | .spec.template.spec.containers[0].env[9].value) = env(LATEST_HASH)' migrator.yaml -i
kubectl -ndevtroncd apply -f migrator.yaml
# shellcheck disable=SC2046
while [ ! $(kubectl -n devtroncd get job postgresql-migrate-devtron -o jsonpath="{.status.succeeded}")  ]; do sleep 10; done
# shellcheck disable=SC2046
while [ ! $(kubectl -n devtroncd get job postgresql-migrate-casbin -o jsonpath="{.status.succeeded}")  ]; do sleep 10; done
# shellcheck disable=SC2046
while [ ! $(kubectl -n devtroncd get job postgresql-migrate-lens -o jsonpath="{.status.succeeded}")  ]; do sleep 10; done
# shellcheck disable=SC2046
while [ ! $(kubectl -n devtroncd get job postgresql-migrate-gitsensor -o jsonpath="{.status.succeeded}")  ]; do sleep 10; done
exit #to get out of container