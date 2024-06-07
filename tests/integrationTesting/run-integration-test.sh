#
# Copyright (c) 2024. Devtron Inc.
#

wget https://go.dev/dl/go1.22.1.linux-amd64.tar.gz -O go1.22.1.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.22.1.tar.gz
export GOPATH='/usr/local/go'
export PATH=$PATH:$GOPATH/bin
#go test ./pkg/pipeline
go test -v ./pkg/pipeline -run TestExecuteWorkflow
exit #to get out of container


