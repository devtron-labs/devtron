INSERT INTO "plugin_metadata" ("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_metadata'), 'Semgrep','Semgrep is a fast, open source, static analysis engine for finding bugs, detecting dependency vulnerabilities, and enforcing code standards.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/semgrep.png','f', 'now()', 1, 'now()', 1);

INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), 2, (SELECT id FROM plugin_metadata WHERE name='Semgrep'),'now()', 1, 'now()', 1);
INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), 3, (SELECT id FROM plugin_metadata WHERE name='Semgrep'),'now()', 1, 'now()', 1);

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
SemgrepTokenLen=$(echo -n $SEMGREP_APP_TOKEN | wc -m)
if [ $((SemgrepTokenLen)) == 0 ]
then
    SEMGREP_APP_TOKEN=$SEMGREP_API_TOKEN
fi
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
    semgrep ci --json $ExtraCommandArguments
done'
        ,
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Semgrep'),'Step 1','Step 1 - Dependency Track for Semgrep','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Semgrep' and ps."index"=1 and ps.deleted=false), 'SemgrepAppToken','STRING','If provided, this token will be used. If not provided it will be picked from secret.',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Semgrep' and ps."index"=1 and ps.deleted=false), 'PrefixAppNameInSemgrepBranchName','BOOL','if true, this will add app name with branch name: {SemgrepAppName}-{branchName}.',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Semgrep' and ps."index"=1 and ps.deleted=false), 'UseCommitAsSemgrepBranchName','BOOL','if true, this will add app name with commit hash: {SemgrepAppName}-{CommitHash}',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Semgrep' and ps."index"=1 and ps.deleted=false), 'SemgrepAppName','STRING','if provided and PrefixAppNameInSemgrepBranchName is true, then this will be prefixed with branch name/ commit hash',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Semgrep' and ps."index"=1 and ps.deleted=false), 'ExtraCommandArguments','STRING','Extra Command arguments for semgrep CI command. eg input: --json --dry-run.',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Semgrep' and ps."index"=1 and ps.deleted=false), 'GIT_MATERIAL_REQUEST','STRING','git material data',false,true,3,'INPUT','GLOBAL',1 ,'GIT_MATERIAL_REQUEST','f','now()', 1, 'now()', 1);

