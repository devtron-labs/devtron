INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Dependency track for Maven & Gradle' and ps."index"=1 and ps.deleted=false),'SkipBuildTest','BOOL','If Enable, this will skips compiling the tests i.e. it skips building the test artifacts','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

UPDATE plugin_pipeline_script SET script=E'mkdir $HOME/outDTrack
OutDirDTrack=$HOME/outDTrack
cd /devtroncd/$CheckoutPath
ToUploadBom=YES

# Convert SkipBuildTest to lowercase for case-insensitive comparison
SkipBuildTestLower=$(echo "$SkipBuildTest" | tr "[:upper:]" "[:lower:]")

if [ $BuildToolType == "GRADLE" ]
then
    if [ "$SkipBuildTestLower" == "true" ]
    then
        apk add gradle
        gradle cyclonedxBom --exclude-task test
        cp build/reports/bom.json $OutDirDTrack/bom.json
    else
        apk add gradle
        gradle cyclonedxBom
        cp build/reports/bom.json $OutDirDTrack/bom.json
    fi
elif [ $BuildToolType == "MAVEN" ]
then
    if [ "$SkipBuildTestLower" == "true" ]
    then
        apk add maven
        mvn install -Dmaven.test.skip=true
        cp target/bom.json $OutDirDTrack/bom.json
    else
        apk add maven
        mvn install
        cp target/bom.json $OutDirDTrack/bom.json
    fi
else
    echo "BUILD_TYPE: $BuildToolType not supported"
    ToUploadBom=NO
fi

if [ $ToUploadBom == "YES" ]
then
    apk add curl
    cd $OutDirDTrack
    curl -v --location --request POST "$DTrackEndpoint/api/v1/bom" \
        --header "accept: application/json" \
        --header "X-Api-Key: $DTrackApiKey" \
        --form "projectName=$DTrackProjectName" \
        --form \"autoCreate=true\" \
        --form "projectVersion=$DTrackProjectVersion" \
        --form "bom=@\"bom.json\""
fi' WHERE id=(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Dependency track for Maven & Gradle' and ps."index"=1 and ps.deleted=false);
