INSERT INTO "plugin_metadata" ("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_metadata'), 'Semgrep','Semgrep is a fast, open source, static analysis engine for finding bugs, detecting dependency vulnerabilities, and enforcing code standards.','SHARED','link_to_icon','f', 'now()', 1, 'now()', 1);

INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), 2, 7,'now()', 1, 'now()', 1);
INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), 3, 7,'now()', 1, 'now()', 1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
    nextval('id_seq_plugin_pipeline_script'),
    '#!/bin/sh
set -eo pipefail
chmod 741 /devtroncd
chmod 741 /devtroncd/*
apk add py3-pip
pip install pip==21.3.1
pip install semgrep
export SEMGREP_APP_TOKEN=$SemgrepAppToken

CiMaterialsEnv=$GIT_MATERIAL_REQUEST
repoName=""
checkoutPath=""
branchName=""
gitHash=""
materials=$(echo $CiMaterialsEnv | tr "|" "\n")
for material in $materials
do
    data=$(echo $material | tr "," "\n")
    i=0
    for d in $data
    do
        if [ $((i)) == 0 ]
        then
            repoName=$d
        elif [ $((i)) == 1 ]
        then
            checkoutPath=$d
        elif [ $((i)) == 2 ]
        then
            branchName=$d
        elif [ $((i)) == 3 ]
        then
            gitHash=$d
        fi
        i=$((i+1))
    done
    #docker run --rm --env SEMGREP_APP_TOKEN=$SemgrepAppToken --env SEMGREP_REPO_NAME=$repoName --env SEMGREP_BRANCH=$branchName -v "${PWD}/:/src/" returntocorp/semgrep semgrep ci
    cd /devtroncd
    cd $checkoutPath
    export SEMGREP_REPO_NAME=$repoName
    if [ $UseCommitAsSemgrepBranchName == true -a $PrefixAppNameInSemgrepBranchName == true ]
    then
        export SEMGREP_BRANCH="$SemgrepAppName - $gitHash"
    elif [ $PrefixAppNameInSemgrepBranchName == true ]
    then
        export SEMGREP_BRANCH="$SemgrepAppName - $branchName"
    elif [ $UseCommitAsSemgrepBranchName == true ]
    then
        export SEMGREP_BRANCH=$gitHash
    else
        export SEMGREP_BRANCH=$branchName
    fi
    semgrep ci $ExtraCommandArguments
done'
        ,
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), 7,'Step 1','Step 1 - Dependency Track for Semgrep','1','INLINE','14','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), 7, 'SemgrepAppToken','STRING','App token of Semgrep account',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), 7, 'PrefixAppNameInSemgrepBranchName','BOOL','if true, this will publish scan results by name {SemgrepAppName}-{branchName}',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), 7, 'UseCommitAsSemgrepBranchName','BOOL','if true, this will publish scan results by name {SemgrepAppName}-{CommitHash}',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), 7, 'SemgrepAppName','STRING','App Name will be used as an extra metadata for publishing results',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), 7, 'ExtraCommandArguments','STRING','Extra Command arguments for semgrep CI command. eg input - --json --sem',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), 7, 'GIT_MATERIAL_REQUEST','STRING','git material data',false,true,3,'INPUT','GLOBAL',1 ,'GIT_MATERIAL_REQUEST','f','now()', 1, 'now()', 1);

