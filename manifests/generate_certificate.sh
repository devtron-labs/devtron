#!/bin/bash

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

set -o errexit
set -o nounset
#set -o pipefail

mkdir -p /tmp/certs

# CREATE THE PRIVATE KEY FOR OUR CUSTOM CA
openssl genrsa -out /tmp/certs/ca.key 2048

# GENERATE A CA CERT WITH THE PRIVATE KEY
openssl req -new -x509 -key /tmp/certs/ca.key -out /tmp/certs/ca.crt -config /tmp/certs/ca_config.txt

# CREATE THE PRIVATE KEY FOR OUR GRUMPY SERVER
openssl genrsa -out /tmp/certs/grumpy-key.pem 2048

# CREATE A CSR FROM THE CONFIGURATION FILE AND OUR PRIVATE KEY
openssl req -new -key /tmp/certs/grumpy-key.pem -subj "/CN=guard-service.devtroncd.svc" -out /tmp/grumpy.csr -config /tmp/certs/grumpy_config.txt

# CREATE THE CERT SIGNING THE CSR WITH THE CA CREATED BEFORE
openssl x509 -req -in /tmp/grumpy.csr -CA /tmp/certs/ca.crt -CAkey /tmp/certs/ca.key -CAcreateserial -out /tmp/certs/grumpy-crt.pem

# INJECT CA IN THE WEBHOOK CONFIGURATION
# export CA_BUNDLE=$(cat certs/ca.crt | base64 | tr -d '\n')
#cat _manifest_.yaml | envsubst > manifest.yaml
