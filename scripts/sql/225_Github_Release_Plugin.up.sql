INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Github Release Plugin v1.0','Managing Releases and tags','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/GithubReleasePluginlogo.png',false,'now()',1,'now()',1);

INSERT INTO plugin_tag (id,name,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_tag'),'Github',false,'now()',1,'now()',1);

INSERT INTO plugin_tag (id,name,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_tag'),'Release',false,'now()',1,'now()',1);

INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), (SELECT id FROM plugin_tag WHERE name='Github'), (SELECT id FROM plugin_metadata WHERE name='Github Release Plugin v1.0'),'now()', 1, 'now()', 1);
INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), (SELECT id FROM plugin_tag WHERE name='Release'), (SELECT id FROM plugin_metadata WHERE name='Github Release Plugin v1.0'),'now()', 1, 'now()', 1);

INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)VALUES (nextval('id_seq_plugin_stage_mapping'),
(SELECT id from plugin_metadata where name='Github Release Plugin v1.0'), 0,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES ( nextval('id_seq_plugin_pipeline_script'),
E'#!/bin/sh
set -e
source_type=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].sourceType\')
# Check if sourceType is "SOURCE_TYPE_BRANCH_FIXED" 
if [ "$source_type" == "SOURCE_TYPE_BRANCH_FIXED" ]; then
    echo "Source type is Branch. Running Plugin..."

    git_repository=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].gitRepository\')
    TARGET=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].sourceValue\')

    if [[ $git_repository =~ ^git@ ]]; then
        username=$(echo "$git_repository" | cut -d \':\' -f 2 | cut -d \'/\' -f 1)
        repo_name=$(echo "$git_repository" | cut -d \'/\' -f 2 | cut -d \'.\' -f 1)
    elif [[ $git_repository =~ ^https:// ]]; then
        username=$(echo "$git_repository" | awk -F/ \'{print $4}\')
        repo_name=$(echo "$git_repository" | awk -F/ \'{print $5}\' | cut -d \'.\' -f 1)
    else
        echo "Invalid Git URL format."
    fi


    auth_token=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].gitOptions.password\')

    if [[ -z $GithubReleaseToken ]]; then
        token=$auth_token
    elif [[ -z $auth_token ]]; then 
        token=$GithubReleaseToken
    else
        echo "Token not found"
    fi

    # Install gh command

    output=$(curl -sS https://webi.sh/gh | sh)
    source ~/.config/envman/PATH.env > /dev/null 2>&1

    echo "$token" | gh auth login --with-token

    if [ ! -f "$GithubReleaseNotesFile" ]; then
        echo "Release notes file not provided."
    fi

    if [ ! -d "$GithubReleaseUploadFolder" ]; then
        echo "Release assets folder not provided."
    fi

    docker_image_tag=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.dockerImageTag\')
    
    GithubReleaseSourcePriority=$(echo "$GithubReleaseSourcePriority" | tr \'[:upper:]\' \'[:lower:]\')
    tag=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.extraEnvironmentVariables.GithubTag\')
  
    if [ -n "$tag" ] && [ "$tag" != "null" ]; then 
        ReleaseTag="$tag"
    else
        if [ "$GithubReleaseSourcePriority"  == "devtronimagetag" ]; then
             ReleaseTag="$docker_image_tag"
        elif [ "$GithubReleaseSourcePriority" == "githubreleasenamepath" ]; then
            ReleaseTag=$(cat "$GithubReleaseNamePath")
        elif [ "$GithubReleaseSourcePriority" == "githubreleasetag" ]; then
             ReleaseTag="$GithubReleaseTag"
        else
          echo "Priority list is not mentioned"
        fi
     fi

    # Check if GithubTag is empty

    if [ -z "$ReleaseTag" ]; then
        echo "GithubTag is empty."
        exit 1
    fi 
        # Prepend user provided prefix if any
    if [ -n "$GithubReleasePrefix" ]; then
            ReleaseTag="${GithubReleasePrefix}-${ReleaseTag}"
    fi

        # Check if tag already exists
        if ! response=$(gh api -X GET repos/$username/$repo_name/git/refs/tags/"${ReleaseTag}" 2>/dev/null); then
            echo "Tag ${ReleaseTag} does not exist. Creating tag."
            var=$(git rev-parse HEAD)
            gh api repos/$username/$repo_name/git/refs -F ref="refs/tags/${ReleaseTag}" -F sha="$var"
        else
            echo "Tag ${ReleaseTag} already exists"
        fi

        if ! gh release view "$ReleaseTag" &>/dev/null; then
                echo "Creating release $ReleaseTag"
                GithubReleaseURL=$(gh release create --target "$TARGET" --title "$ReleaseTag" ${GithubReleaseNotesFile:+--notes-file "$GithubReleaseNotesFile"} "$ReleaseTag" --verify-tag)
                 success_message="Github Tag $ReleaseTag and Release $ReleaseTag successfully created."
            if [ -n "$GithubReleaseUploadFolder" ]; then  
                # Upload release assets if the release was created
                for file in "$GithubReleaseUploadFolder"/*; do
                    if [ -f "$file" ]; then 
                        gh release upload "$ReleaseTag" "$file"
                        echo "Release assets uploaded successfully."
                    fi
                done
            else
                echo "No release assets provided."
            fi
        else
            echo "Release ${ReleaseTag} already exist"  
            if [ -n "$GithubReleaseUploadFolder" ]; then  
                # Upload release assets if the release was created
                for file in "$GithubReleaseUploadFolder"/*; do
                    if [ -f "$file" ]; then 
                        gh release upload "$ReleaseTag" "$file"
                    fi
                done
                echo "Release assets uploaded successfully."

            else
                echo "No release assets provided."
            fi
        fi
      
 export GithubReleaseURL
echo "Release link: $GithubReleaseURL"
if [ -n "$success_message" ]; then
    echo "$success_message"
fi

else
    echo "Source type is not Branch_fixed. Cannot run the plugin"
fi',
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Github Release Plugin v1.0'),'Step 1','Step 1 - Github Release Plugin v1.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES 
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Release Plugin v1.0' and ps."index"=1 and ps.deleted=false),'GithubReleaseToken','STRING','Enter the Github Personal Acess Token(PAT) with the necessary permissions for creating releases and tags.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Release Plugin v1.0' and ps."index"=1 and ps.deleted=false),'GithubReleaseSourcePriority','STRING','Specify the priority for selecting the source from which to create tag and releases.Options in the priority includes: 1.DevtronImageTag(tag associated with a Docker image) 2.GithubReleaseNamepath 3.GithubReleaseTag','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Release Plugin v1.0' and ps."index"=1 and ps.deleted=false),'GithubReleaseTag','STRING','Enter the tag for the release.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Release Plugin v1.0' and ps."index"=1 and ps.deleted=false),'GithubReleaseNamePath','STRING','Enter the path to the file containing Tag and Release name.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Release Plugin v1.0' and ps."index"=1 and ps.deleted=false),'GithubReleasePrefix','STRING','Enter the prefix to be added on tag and release name.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Release Plugin v1.0' and ps."index"=1 and ps.deleted=false),'GithubReleaseNotesFile','STRING','Enter the path to the Release notes file.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Release Plugin v1.0' and ps."index"=1 and ps.deleted=false),'GithubReleaseUploadFolder','STRING','Enter the path to the folder containing Release assets.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Release Plugin v1.0' and ps."index"=1 and ps.deleted=false),'GithubReleaseURL','STRING','Release URL contains all relevant information about the release, including assets and release notes.','t','f',null,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);




