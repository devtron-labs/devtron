git clone https://github.com/devtron-labs/devtron.git
cd devtron || exit
git checkout $TEST_BRANCH
wget https://go.dev/dl/go1.19.10.linux-amd64.tar.gz -O go1.19.10.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.10.tar.gz
export PATH=$PATH:/usr/local/go/bin
go test ./tests/integrationTesting
exit #to get out of container


