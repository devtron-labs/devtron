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

wget https://go.dev/dl/go1.22.1.linux-amd64.tar.gz -O go1.22.1.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.22.1.tar.gz
export GOPATH='/usr/local/go'
export PATH=$PATH:$GOPATH/bin
#go test ./pkg/pipeline
go test -v ./pkg/pipeline -run TestExecuteWorkflow
exit #to get out of container


