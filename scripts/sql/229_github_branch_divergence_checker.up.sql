/*
 * Copyright (c) 2024. Devtron Inc.
 */

INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'GitHub Branch Divergence Checker v1.0','Comparing Commit Histories in Branches in Github Repository','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/branch-compare-plugin-logo.png',false,'now()',1,'now()',1);

INSERT INTO plugin_tag_relation ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), (SELECT id FROM plugin_tag WHERE name='Github'), (SELECT id FROM plugin_metadata WHERE name='GitHub Branch Divergence Checker v1.0'),'now()', 1, 'now()', 1);


INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)VALUES (nextval('id_seq_plugin_stage_mapping'),
(SELECT id from plugin_metadata where name='GitHub Branch Divergence Checker v1.0'), 0,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES ( nextval('id_seq_plugin_pipeline_script'),
E'#!/bin/bash
set -e
source_type=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].sourceType\')
if [ "$source_type" == "WEBHOOK" ]; then
     if [ -z "$ParentBranch" ]; then 
        ParentBranch=master
    fi

    if [ -z "$ChildBranch" ]; then
        ChildBranch=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].WebhookData.data."source branch name"\')
    fi 
   
elif [ "$source_type" == "SOURCE_TYPE_BRANCH_FIXED" ]; then
     if [ -z "$ParentBranch" ]; then  
        ParentBranch=master
    fi
    
    if [ -z "$ChildBranch" ]; then
       ChildBranch=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].sourceValue\') 
    fi
   
else
    echo "Source type provided is not appropriate."
fi

git checkout "$ChildBranch"
branch2=${2:-origin/$ChildBranch}
branch1=${1:-origin/$ParentBranch}

commits_ahead=$(git rev-list --no-merges --count "$branch2".."$branch1")

if [ "$commits_ahead" -gt 2 ]; then
    echo "$branch2 is ahead of $branch1 by $commits_ahead commits. Please merge $branch2 into $branch1 first."
    exit 1
else
    echo "$branch1 and $branch2 branches are in sync or the commit difference is within the allowed range."
fi',      
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='GitHub Branch Divergence Checker v1.0'),'Step 1','Step 1 - GitHub Branch Divergence Checker v1.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES 
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GitHub Branch Divergence Checker v1.0' and ps."index"=1 and ps.deleted=false),'ParentBranch','STRING','Enter the branch that will be compared against the commits of ChildBranch. Default: master, if not provided.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GitHub Branch Divergence Checker v1.0' and ps."index"=1 and ps.deleted=false),'ChildBranch','STRING','Enter the branch that needs to be checked for commits.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

