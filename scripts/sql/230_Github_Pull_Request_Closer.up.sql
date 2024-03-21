INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Github Pull Request Closer v1.0','Closing Pull Requests in Github ','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/GithubReleasePR.png',false,'now()',1,'now()',1);

INSERT INTO plugin_tag_relation ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), (SELECT id FROM plugin_tag WHERE name='Github'), (SELECT id FROM plugin_metadata WHERE name='Github Pull Request Closer v1.0'),'now()', 1, 'now()', 1);

INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)VALUES (nextval('id_seq_plugin_stage_mapping'),
(SELECT id from plugin_metadata where name='Github Pull Request Closer v1.0'), 0,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES ( nextval('id_seq_plugin_pipeline_script'),
E'#!/bin/bash
set -e
    source_type=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].sourceType\')
     if [ "$source_type" == "WEBHOOK" ]; then
        echo "Source type is Pull Request. Running script..."

        grep_option="Fqe" 
         if [ -z "$GrepCommand" ]; then
        GrepCommand="$grep_option"
    fi  
     if echo "$PreviousStepOutputVariable" | grep -"$GrepCommand" "$PreviousStepOutputGrepPattern" 
     then  
          echo "Pattern Matched. Running the plugin script..." 
          git_repository=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].gitRepository\')
          if [[ $git_repository =~ ^git@ ]]; then
                username=$(echo "$git_repository" | cut -d \':\' -f 2 | cut -d \'/\' -f 1)
                repo_name=$(echo "$git_repository" | cut -d \'/\' -f 2 | cut -d \'.\' -f 1)
        elif [[ $git_repository =~ ^https:// ]]; then
            username=$(echo "$git_repository" | awk -F/ \'{print $4}\')
            repo_name=$(echo "$git_repository" | awk -F/ \'{print $5}\' | cut -d \'.\' -f 1)
        else
            echo "Invalid Git URL format."
        fi

        token=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].gitOptions.password\')
        repo_url=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].WebhookData.data."git url"\')
        owner_repo=$(echo "$repo_url" | awk -F/ \'{print $(NF-3)"/"$(NF-2)}\')
        number=$(echo "$repo_url" | awk -F/ \'{print $(NF)}\')
        
        output=$(curl -sS https://webi.sh/gh | sh)
        source ~/.config/envman/PATH.env > /dev/null 2>&1

        echo "$token" | gh auth login --with-token
        pr_state=$(gh pr view "$number" --repo "$owner_repo" --json state | jq -r \'.state\')

        if [ "$pr_state" == "closed" ]; then
            echo "Pull request is already closed."
        else
            if [ -z "$GithubPRCloseComment" ]; then
                 echo "No comment provided. Closing PR without commenting..."
                if gh pr close "$repo_url"; then
                    echo "Pull request closed successfully."
                else
                    echo "Failed to close the pull request."
                    exit 1
                fi
            else
                 echo "Comment provided. Commenting on PR..." 
                 gh pr comment "$number" --body "$GithubPRCloseComment"
                  if gh pr close "$repo_url"; then 
                     echo "Pull request closed successfully."
                  else
                     echo "Failed to close the pull request."
                     exit 1
                fi
            fi
        fi
        
    else
         echo "Pattern does not match" 
    fi

    else
        echo "Source type is not Pull Request. Skipping plugin execution."
        exit1
    fi',
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Github Pull Request Closer v1.0'),'Step 1','Step 1 - Github Pull Request Closer v1.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES 
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Pull Request Closer v1.0' and ps."index"=1 and ps.deleted=false),'PreviousStepOutputVariable','STRING','Use the output variable obtained from the last script execution.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Pull Request Closer v1.0' and ps."index"=1 and ps.deleted=false),'PreviousStepOutputGrepPattern','STRING',' Enter the pattern or value to be compared to search in the previousStepOutputVariable using the grep command','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Pull Request Closer v1.0' and ps."index"=1 and ps.deleted=false),'GrepCommand','STRING','Enter the command options to be used with grep. Default Command:"Fqe", if not provided.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Pull Request Closer v1.0' and ps."index"=1 and ps.deleted=false),'GithubPRCloseComment','STRING','Enter the comment that should be written when closing the pull request on GitHub.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

