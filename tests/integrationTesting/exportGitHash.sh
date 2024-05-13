GITHASH=$GIT_HASH
if [ -z "$GITHASH" ]; then
    echo "-------laeeeeeeeeeeeeq-------------"
    git log --pretty=format:'%h' -n 1
    echo "------laeeqenccddddd----------------"
fi;
echo "GIT_HASH is" $GITHASH
docker exec dind-test sh -c "export GIT_HASH=$GITHASH && mkdir test && cp -r wirenil/* test/ && ./test/tests/integrationTesting/exportEnvsExecuteWireNilChecker.sh"