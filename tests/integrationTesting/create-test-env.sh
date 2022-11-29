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
yq '((select .spec != null) | .spec.template.spec.containers[0] | select (.name == "postgresql-migrate-devtron" and .name != null) | .env[0].value) = env(TEST_BRANCH)' migrator.yaml -i
yq '(.spec.template.spec.containers[0] | select (.name == "postgresql-migrate-devtron") | .env[9].value) = env(LATEST_HASH)' migrator.yaml -i
kubectl apply -f migrator.yaml
exit #to get out of container