#
# Copyright (c) 2024. Devtron Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
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