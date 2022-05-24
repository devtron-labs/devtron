INSERT INTO "public"."plugin_tag" ("id", "name", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_tag'), 'DevSecOps', 'f', 'now()', '1', 'now()', '1');


--- dTrack plugin for python

INSERT INTO "public"."plugin_metadata" ("id", "name", "description", "type", "icon", "deleted", "created_on",
                                        "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_metadata'), 'Dependency track for Python',
        'Creates a bill of materials from Python projects and environments and uploads it to D-track for Component Analysis, to identify and reduce risk in the software supply chain.',
        'PRESET',
        'https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/dTrack-plugin-icon.png', 'f', 'now()', '1',
        'now()', '1');


INSERT INTO "public"."plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on",
                                            "updated_by")
VALUES (nextval('id_seq_plugin_tag_relation'), 3,
        (select currval('id_seq_plugin_metadata')), 'now()', '1', 'now()', '1'),
       (nextval('id_seq_plugin_tag_relation'), (select currval('id_seq_plugin_tag')),
        (select currval('id_seq_plugin_metadata')), 'now()', '1', 'now()', '1');



INSERT INTO "public"."plugin_pipeline_script" ("id", "script", "type", "deleted", "created_on", "created_by",
                                               "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_pipeline_script'), 'apk add py3-pip
pip install cyclonedx-bom
mkdir $HOME/outDTrack
OutDirDTrack=$HOME/outDTrack
cd /devtroncd/$CheckoutPath
ToUploadBom=YES
if [ $ProjectManifestType == "POETRY" ]
then
	cyclonedx-bom -i $RelativePathToPoetryLock -o $OutDirDTrack/bom.json --format json -p
elif [ $ProjectManifestType == "PIP" ]
then
	cyclonedx-bom -i $RelativePathToPipfile -o $OutDirDTrack/bom.json --format json -pip
elif [ $ProjectManifestType == "REQUIREMENT" ]
then
	cyclonedx-bom -i $RelativePathToRequirementTxt -o $OutDirDTrack/bom.json --format json -r
elif [ $ProjectManifestType == "ENV" ]
then
	cyclonedx-bom -o $OutDirDTrack/bom.json --format json -e
else
    echo "MANIFEST_TYPE: $ProjectManifestType not supported"
    ToUploadBom=NO
fi

if [ $ToUploadBom == "YES" ]
then
	apk add curl
    cd $OutDirDTrack
	curl -v --location --request POST "$DTrackEndpoint/api/v1/bom" \
   		--header ''accept: application/json'' \
    	--header "X-Api-Key: $DTrackApiKey" \
    	--form "projectName=$DTrackProjectName" \
    	--form ''autoCreate="true"'' \
    	--form "projectVersion=$DTrackProjectVersion" \
    	--form ''bom=@"bom.json"''
fi', 'SHELL', 'f', 'now()', '1', 'now()', '1');


INSERT INTO "public"."plugin_step" ("id", "plugin_id", "name", "description", "index", "step_type", "script_id",
                                    "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES ((nextval('id_seq_plugin_step')), (select currval('id_seq_plugin_metadata')), 'Step 1',
        'Step 1 - Dependency Track for Python)', '1', 'INLINE', (select currval('id_seq_plugin_pipeline_script')), 'f',
        'now()', '1', 'now()', '1');


INSERT INTO "public"."plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed",
                                             "allow_empty_value", "variable_type", "value_type", "default_value",
                                             "variable_step_index", "deleted", "created_on", "created_by", "updated_on",
                                             "updated_by")
VALUES ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'ProjectManifestType',
        'STRING',
        'type of your python project manifest which is to be used to build cycloneDx SBOM. OneOf - PIP, POETRY, ENV, REQUIREMENT',
        't', 'f',
        'INPUT', 'NEW', 'ENV', '1', 'f', 'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')),
        'RelativePathToPoetryLock',
        'STRING', 'Path to your poetry.lock file inside your project.', 't', 't', 'INPUT', 'NEW', 'poetry.lock', '1',
        'f', 'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')),
        'RelativePathToPipfile',
        'STRING', 'Path to your Pipfile.lock file inside your project.', 't', 't', 'INPUT', 'NEW', 'Pipfile.lock', '1',
        'f', 'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')),
        'RelativePathToRequirementTxt',
        'STRING', 'Path to your requirements.txt file inside your project.', 't', 't', 'INPUT', 'NEW',
        'requirements.txt', '1',
        'f', 'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'DTrackEndpoint',
        'STRING', 'Api endpoint of your dependency track account.', 't', 'f', 'INPUT', 'NEW', NULL, '1', 'f',
        'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'DTrackProjectName',
        'STRING', 'Name of dependency track project.', 't', 'f', 'INPUT', 'NEW', NULL, '1', 'f',
        'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'DTrackProjectVersion',
        'STRING', 'Version of dependency track project.', 't', 'f', 'INPUT', 'NEW', NULL, '1', 'f',
        'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'DTrackApiKey',
        'STRING', 'Api key of your dependency track account.', 't', 'f', 'INPUT', 'NEW', NULL, '1', 'f',
        'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'CheckoutPath',
        'STRING', 'Checkout path of git material.', 't', 'f', 'INPUT', 'NEW', './', '1', 'f',
        'now()', '1', 'now()', '1');


--- dTrack plugin for node.js

INSERT INTO "public"."plugin_metadata" ("id", "name", "description", "type", "icon", "deleted", "created_on",
                                        "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_metadata'), 'Dependency track for NodeJs',
        'Creates a bill of materials from NodeJs projects and environments and uploads it to D-track for Component Analysis, to identify and reduce risk in the software supply chain.',
        'PRESET',
        'https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/dTrack-plugin-icon.png', 'f', 'now()', '1',
        'now()', '1');


INSERT INTO "public"."plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on",
                                            "updated_by")
VALUES (nextval('id_seq_plugin_tag_relation'), 3,
        (select currval('id_seq_plugin_metadata')), 'now()', '1', 'now()', '1'),
       (nextval('id_seq_plugin_tag_relation'), (select currval('id_seq_plugin_tag')),
        (select currval('id_seq_plugin_metadata')), 'now()', '1', 'now()', '1');



INSERT INTO "public"."plugin_pipeline_script" ("id", "script", "type", "deleted", "created_on", "created_by",
                                               "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_pipeline_script'), 'apk add npm
npm install -g @cyclonedx/bom
mkdir $HOME/outDTrack
OutDirDTrack=$HOME/outDTrack
cd /devtroncd/$CheckoutPath
npm install
cyclonedx-node -o $OutDirDTrack/bom.json
apk add curl
cd $OutDirDTrack
curl -v --location --request POST "$DTrackEndpoint/api/v1/bom" \
	--header ''accept: application/json'' \
	--header "X-Api-Key: $DTrackApiKey" \
	--form "projectName=$DTrackProjectName" \
	--form ''autoCreate="true"'' \
	--form "projectVersion=$DTrackProjectVersion" \
	--form ''bom=@"bom.json"''', 'SHELL', 'f', 'now()', '1', 'now()', '1');


INSERT INTO "public"."plugin_step" ("id", "plugin_id", "name", "description", "index", "step_type", "script_id",
                                    "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES ((nextval('id_seq_plugin_step')), (select currval('id_seq_plugin_metadata')), 'Step 1',
        'Step 1 - Dependency Track for NodeJs', '1', 'INLINE', (select currval('id_seq_plugin_pipeline_script')), 'f',
        'now()', '1', 'now()', '1');


INSERT INTO "public"."plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed",
                                             "allow_empty_value", "variable_type", "value_type", "default_value",
                                             "variable_step_index", "deleted", "created_on", "created_by", "updated_on",
                                             "updated_by")
VALUES ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'DTrackEndpoint',
        'STRING', 'Api endpoint of your dependency track account.', 't', 'f', 'INPUT', 'NEW', NULL, '1', 'f',
        'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'DTrackProjectName',
        'STRING', 'Name of dependency track project.', 't', 'f', 'INPUT', 'NEW', NULL, '1', 'f',
        'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'DTrackProjectVersion',
        'STRING', 'Version of dependency track project.', 't', 'f', 'INPUT', 'NEW', NULL, '1', 'f',
        'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'DTrackApiKey',
        'STRING', 'Api key of your dependency track account.', 't', 'f', 'INPUT', 'NEW', NULL, '1', 'f',
        'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'CheckoutPath',
        'STRING', 'Checkout path of git material.', 't', 'f', 'INPUT', 'NEW', './', '1', 'f',
        'now()', '1', 'now()', '1');


--- dTrack plugin for maven & gradle

INSERT INTO "public"."plugin_metadata" ("id", "name", "description", "type", "icon", "deleted", "created_on",
                                        "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_metadata'), 'Dependency track for Maven & Gradle)',
        'Creates a bill of materials from Maven/Gradle projects and environments and uploads it to D-track for Component Analysis, to identify and reduce risk in the software supply chain.',
        'PRESET',
        'https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/dTrack-plugin-icon.png', 'f', 'now()', '1',
        'now()', '1');


INSERT INTO "public"."plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on",
                                            "updated_by")
VALUES (nextval('id_seq_plugin_tag_relation'), 3,
        (select currval('id_seq_plugin_metadata')), 'now()', '1', 'now()', '1'),
       (nextval('id_seq_plugin_tag_relation'), (select currval('id_seq_plugin_tag')),
        (select currval('id_seq_plugin_metadata')), 'now()', '1', 'now()', '1');



INSERT INTO "public"."plugin_pipeline_script" ("id", "script", "type", "deleted", "created_on", "created_by",
                                               "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_pipeline_script'), 'mkdir $HOME/outDTrack
OutDirDTrack=$HOME/outDTrack
cd /devtroncd/$CheckoutPath
ToUploadBom=YES
if [ $BuildToolType == "GRADLE" ]
then
	apk add gradle
	gradle cyclonedxBom
	cp build/reports/bom.json $OutDirDTrack/bom.json
elif [ $BuildToolType == "MAVEN" ]
then
	apk add maven
	mvn install
	cp target/bom.json $OutDirDTrack/bom.json
else
    echo "BUILD_TYPE: $BuildToolType not supported"
    ToUploadBom=NO
fi

if [ $ToUploadBom == "YES" ]
then
	apk add curl
    cd $OutDirDTrack
	curl -v --location --request POST "$DTrackEndpoint/api/v1/bom" \
   		--header ''accept: application/json'' \
    	--header "X-Api-Key: $DTrackApiKey" \
    	--form "projectName=$DTrackProjectName" \
    	--form ''autoCreate="true"'' \
    	--form "projectVersion=$DTrackProjectVersion" \
    	--form ''bom=@"bom.json"''
fi', 'SHELL', 'f', 'now()', '1', 'now()', '1');


INSERT INTO "public"."plugin_step" ("id", "plugin_id", "name", "description", "index", "step_type", "script_id",
                                    "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES ((nextval('id_seq_plugin_step')), (select currval('id_seq_plugin_metadata')), 'Step 1',
        'Step 1 for Dependency Track for Maven & Gradle)', '1', 'INLINE',
        (select currval('id_seq_plugin_pipeline_script')),
        'f',
        'now()', '1', 'now()', '1');


INSERT INTO "public"."plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed",
                                             "allow_empty_value", "variable_type", "value_type", "default_value",
                                             "variable_step_index", "deleted", "created_on", "created_by", "updated_on",
                                             "updated_by")
VALUES ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'BuildToolType',
        'STRING', 'Type of build tool your project is using. OneOf - MAVEN, GRADLE.', 't', 'f', 'INPUT', 'NEW', NULL,
        '1', 'f',
        'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'DTrackEndpoint',
        'STRING', 'Api endpoint of your dependency track account.', 't', 'f', 'INPUT', 'NEW', NULL, '1', 'f',
        'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'DTrackProjectName',
        'STRING', 'Name of dependency track project.', 't', 'f', 'INPUT', 'NEW', NULL, '1', 'f',
        'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'DTrackProjectVersion',
        'STRING', 'Version of dependency track project.', 't', 'f', 'INPUT', 'NEW', NULL, '1', 'f',
        'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'DTrackApiKey',
        'STRING', 'Api key of your dependency track account.', 't', 'f', 'INPUT', 'NEW', NULL, '1', 'f',
        'now()', '1', 'now()', '1'),
       ((nextval('id_seq_plugin_step_variable')), (select currval('id_seq_plugin_metadata')), 'CheckoutPath',
        'STRING', 'Checkout path of git material.', 't', 'f', 'INPUT', 'NEW', './', '1', 'f',
        'now()', '1', 'now()', '1');