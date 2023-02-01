git clone https://github.com/devtron-labs/devtron.git
cd devtron || exit
git checkout $TEST_BRANCH
go test ./tests/integrationTesting
exit #to get out of container


