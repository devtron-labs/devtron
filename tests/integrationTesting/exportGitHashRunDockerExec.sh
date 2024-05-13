GITHASH=$GIT_HASH
if [ -z "$GITHASH" ]; then
    export GITHASH=$(git log --pretty=format:'%h' -n 1)
fi;
docker exec dind-test sh -c "export GIT_HASH=$GITHASH && mkdir test && cp -r wirenil/* test/ && ./test/tests/integrationTesting/exportEnvsExecuteWireNilChecker.sh"