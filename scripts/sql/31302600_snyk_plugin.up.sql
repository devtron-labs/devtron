INSERT INTO "plugin_parent_metadata" ("id", "name","identifier", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_parent_metadata'), 'Code Scan from Snyk','snyk-scan','Scans the code for the vulnerabilities','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/plugin-icons/ic-plugin-snyk-scan.png','f', 'now()', 1, 'now()', 1);


UPDATE plugin_metadata SET is_latest = false WHERE id = (SELECT id FROM plugin_metadata WHERE name= 'Code Scan from Snyk' and is_latest= true);


INSERT INTO "plugin_metadata" ("id", "name", "description","deleted", "created_on", "created_by", "updated_on", "updated_by","plugin_parent_metadata_id","plugin_version","is_deprecated","is_latest")
VALUES (nextval('id_seq_plugin_metadata'), 'Code Scan from Snyk','Update the configurations for the environment','f', 'now()', 1, 'now()', 1, (SELECT id FROM plugin_parent_metadata WHERE identifier='snyk-scan'),'1.0.0', false, true);


INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE plugin_version='1.0.0' and name='Code Scan from Snyk' and deleted= false),3,'now()', 1, 'now()', 1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")VALUES (
    nextval('id_seq_plugin_pipeline_script'),
    E'#!/bin/sh
if [ -z "$ScanContext" ];then
    build_context=$(echo "$CI_CD_EVENT" | jq -r ".commonWorkflowRequest.ciBuildConfig.dockerBuildConfig.buildContext")
    if [ -z "$build_context" ];then
        build_context=".";
    fi
else
    build_context=$ScanContext
fi


cd $build_context;
echo "Scan context is $PWD"
docker run --rm --env SNYK_TOKEN=$ApiKey -v $PWD:/app $ImageTag
exit_code=$?
if [ "$AbortBuildOnVulnerableCode" = true ];then
    if [ $exit_code = 1 ];then
        exit $exit_code
    fi
else
    if [ $exit_code = 1 ] || [ $exit_code = 0 ];then
        continue;
    else
        exit $exit_code
    fi
fi',
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'),(SELECT id FROM plugin_metadata WHERE plugin_version='1.0.0' and name='Code Scan from Snyk' and deleted= false),'Step 1','Step 1 - Scanning the code','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by","default_value")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='1.0.0' and p.name='Code Scan from Snyk' and p.deleted=false and ps."index"=1 and ps.deleted=false), 'ApiKey','STRING','Provide Snyk API Key of your organization',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1, null),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='1.0.0' and p.name='Code Scan from Snyk' and p.deleted=false and ps."index"=1 and ps.deleted=false), 'ImageTag','STRING','Specify the image tag of the snyk tool to be used while scanning',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1,null),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='1.0.0' and p.name='Code Scan from Snyk' and p.deleted=false and ps."index"=1 and ps.deleted=false), 'AbortBuildOnVulnerableCode','STRING','If set true it will abort the build if scanning found some vulnerabilities in code.',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1,'false'),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='1.0.0' and p.name='Code Scan from Snyk' and p.deleted=false and ps."index"=1 and ps.deleted=false), 'ScanContext','STRING','Specify the context to scan. Default is same as build context',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1,null);