#
# Copyright (c) 2024. Devtron Inc.
#

cd test
./tests/integrationTesting/create-test-env.sh
./tests/integrationTesting/run-integration-test.sh
touch output.env
export NODE_IP_ADDRESS=$(kubectl get node  --no-headers  -o custom-columns=INTERNAL-IP:status.addresses[0].address)
export PG_ADDR=$NODE_IP_ADDRESS
export NATS_SERVER_HOST=nats://$NODE_IP_ADDRESS:30236
go mod tidy
go run .
cp output.env ../tempfile