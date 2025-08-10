UPDATE plugin_pipeline_script SET script=E'#!/bin/bash
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

branch2=${2:-origin/$ChildBranch}
branch1=${1:-origin/$ParentBranch}

commits_ahead=$(git rev-list --no-merges --count "$branch2".."$branch1")

if [ "$commits_ahead" -gt 2 ]; then
    echo "$branch2 is ahead of $branch1 by $commits_ahead commits. Please merge $branch2 into $branch1 first."
    exit 1
else
    echo "$branch1 and $branch2 branches are in sync or the commit difference is within the allowed range."
fi' 
WHERE id=(select script_id  from plugin_step where id=(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GitHub Branch Divergence Checker v1.0' and ps."index"=1 and ps.deleted=false));

