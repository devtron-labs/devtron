docker exec dind-test sh -c "export GIT_HASH=$GITHASH && mkdir test && cp -r wirenil/* test/ && ./test/tests/integrationTesting/exportEnvsExecuteWireNilChecker.sh"