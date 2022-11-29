# shellcheck disable=SC2155
export INTEGRATION_TEST_ENV_ID=$(docker run --env TEST_BRANCH --env LATEST_HASH --privileged -d --name dind-test docker:dind)
docker exec -i dind-test /bin/sh
apk update && apk add wget && apk add curl && apk add vim  && apk add bash && apk add git
wget -q -O - https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
k3d cluster create it-cluster
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
wget https://raw.githubusercontent.com/devtron-labs/devtron/main/tests/integrationTesting/postgresql-secret.yaml
kubectl apply -f postgresql-secret.yaml
wget https://raw.githubusercontent.com/devtron-labs/devtron/main/tests/integrationTesting/postgresql.yaml
kubectl apply -f postgresql.yaml
wget https://raw.githubusercontent.com/devtron-labs/devtron/main/tests/integrationTesting/migrator.yaml
yq '(.spec.template.spec.containers[0] | select (.name == "postgresql-migrate-devtron") | .env[0].value) = env(TEST_BRANCH)' migrator.yaml
yq '(.spec.template.spec.containers[0] | select (.name == "postgresql-migrate-devtron") | .env[9].value) = env(LATEST_HASH)' migrator.yaml
kubectl apply -f migrator.yaml
exit #to get out of container