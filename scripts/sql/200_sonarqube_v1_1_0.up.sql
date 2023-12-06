INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Sonarqube-v1.1.0','Enhance your workflow with continuous code quality and code security using the Sonarqube-v1.1.0 plugin','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/sonarqube-plugin-icon.png',false,'now()',1,'now()',1);


INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), (SELECT id FROM plugin_tag WHERE name='Code Review'), (SELECT id FROM plugin_metadata WHERE name='Sonarqube-v1.1.0'),'now()', 1, 'now()', 1);
INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), (SELECT id FROM plugin_tag WHERE name='Security'), (SELECT id FROM plugin_metadata WHERE name='Sonarqube-v1.1.0'),'now()', 1, 'now()', 1);
INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), (SELECT id FROM plugin_tag WHERE name='Code Review'), (SELECT id FROM plugin_metadata WHERE name='Sonarqube-v1.1.0'),'now()', 1, 'now()', 1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
    nextval('id_seq_plugin_pipeline_script'),
    '#!/bin/sh
    repoName=""
    branchName=""
    # Define the function to extract repoName and branchName
    FetchRepoBranchNameFunction() {
        CiMaterialsRequests=$GIT_MATERIAL_REQUEST
        materials=$(echo $CiMaterialsRequests | tr "|" "\n")
        for material in $materials
        do
            # echo "material : $material"
            data=$(echo $material | tr "," "\n")
            # echo "data: $data"
            repo_name=$(echo "$data" | sed -n ''1p'')
            branch_name=$(echo "$data" | sed -n ''3p'')
            # echo Reponame: $repo_name and branchName: $branch_name
            repoName="${repoName}-$repo_name"
            branchName="${branchName}-$branch_name"
        done
        repoName="${repoName#-}"
        branchName="${branchName#-}"
    }
    GlobalSonarqubeProjectName=""
    GlobalSonarqubeBranchName=""
    # Define sonarqube scan function
    SonarqubeScanFunction() {
    echo -e "\n********** Starting the scanning ************"
    docker run --rm -e SONAR_HOST_URL=$SonarqubeEndpoint -e SONAR_LOGIN=$SonarqubeApiKey -v "/$PWD:/usr/src" sonarsource/sonar-scanner-cli
    SonarScanStatusCode=$?
    echo -e "\nStatus code of sonarqube scanning command : $SonarScanStatusCode"
    if [ "$SonarScanStatusCode" -ne 0 ]; then
        echo -e "****** Sonarqube scanning command failed to run *********"
        exit 1
    fi
    if [[ $CheckForSonarAnalysisReport == true && ! -z "$CheckForSonarAnalysisReport" ]]
        then
        status=$(curl -u ${SonarqubeApiKey}:  -sS ${SonarqubeEndpoint}/api/qualitygates/project_status?projectKey=$GlobalSonarqubeProjectName&branch=$SonarqubeBranchName)
        project_status=$(echo $status | jq -r  ".projectStatus.status")
        export SonarqubeProjectStatus=$project_status
        echo "*********  SonarQube Policy Report  *********"
        echo $status
        if [[ $AbortPipelineOnPolicyCheckFailed == true && $project_status == "ERROR" ]]
        then
        echo "*********  SonarQube Policy Violated *********"
        echo "*********  Exiting Build *********"
        exit
        elif [[ $AbortPipelineOnPolicyCheckFailed == true && $project_status == "OK" ]]
        then
        echo "*********  SonarQube Policy Passed *********"
        fi
    else
        echo -e "\nFinding the Vulnerabilities and High hotspots in source code ........\n"
        sleep 5
        export SonarqubeVulnerabilities=$(curl -u ${SonarqubeApiKey}: --location --request GET "$SonarqubeEndpoint/api/issues/search?componentKeys=$GlobalSonarqubeProjectNamee&types=VULNERABILITY" | jq ".issues | length")
        export SonarqubeHighHotspots=$(curl -u ${SonarqubeApiKey}: --location --request GET "$SonarqubeEndpoint/api/hotspots/search?projectKey=$GlobalSonarqubeProjectName" | jq ''.hotspots|[.[]|select(.vulnerabilityProbability=="HIGH")]|length'')
        echo "Total Sonarqube Vulnerability: $SonarqubeVulnerabilities"
        echo "Total High Hotspots:  $SonarqubeHighHotspots"
        export TotalSonarqubeIssues=$((SonarqubeVulnerabilities + SonarqubeHighHotspots))
        echo "Total number of issues found by sonarqube scanner : $TotalSonarqubeIssues"
        echo -e "For analysis report please visit $SonarqubeEndpoint/dashboard?id=$GlobalSonarqubeProjectName"
    fi
    }


    FetchRepoBranchNameFunction
    if [ -z $SonarqubeProjectPrefixName ]
    then
    SonarqubeProjectPrefixName=$repoName
    fi
    if [ -z $SonarqubeBranchName ]
    then
    SonarqubeBranchName=$branchName
    fi


    PathToCodeDir=/devtroncd$CheckoutPath
    cd $PathToCodeDir
    if [ ! -z $SonarqubeProjectKey ]
    then
    GlobalSonarqubeProjectName=$SonarqubeProjectKey
    GlobalSonarqubeBranchName="master"
    else
    GlobalSonarqubeProjectName=$SonarqubeProjectPrefixName-$SonarqubeBranchName
    GlobalSonarqubeBranchName=$SonarqubeBranchName
    fi
    if [[ -z "$UsePropertiesFileFromProject" || $UsePropertiesFileFromProject == false ]]
    then
    echo "sonar.projectKey=$GlobalSonarqubeProjectName" > sonar-project.properties
    fi
    echo -e "\n********** Sonarqube Project Name : $GlobalSonarqubeProjectName , Sonarqube Branch name : $SonarqubeBranchName ***********"
    if [ -z "$GlobalSonarqubeProjectName" ] || [ -z "$SonarqubeBranchName" ]; then
    echo -e "\n****** Sonarqube Project Name and Sonarqube branch name should not be empty *********"
    exit 1
    fi

    if [ -z $SonarqubeApiKey ]
    then
    echo "************* Sonarqube analysis api key has not been provided *************"
    exit 1
    fi
    if [ -z $SonarqubeEndpoint ]
    then
        echo "********** Sonarqube endpoint URL has not been provided ********* "
        exit 1
    fi

    echo -e "\n*********Creating Sonarqube project **********"
    curl -u ${SonarqubeApiKey}: --location --request POST "$SonarqubeEndpoint/api/projects/create?name=$GlobalSonarqubeProjectName&mainBranch=$SonarqubeBranchName&project=$GlobalSonarqubeProjectName"
    CreateProjectStatusCode=$?
    if [ "$CreateProjectStatusCode" -ne 0 ]; then
    echo -e "****** Sonarqube project create command failed to run *********"
    exit 1
    else
    SonarqubeScanFunction
    fi',
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);
INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Sonarqube-v1.1.0'),'Step 1','Step 1 - Sonarqube v1.1.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube-v1.1.0' and ps."index"=1 and ps.deleted=false),'SonarqubeProjectPrefixName','STRING','This is the SonarQube project prefix name. If not provided, the prefix name is automatically generated.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube-v1.1.0' and ps."index"=1 and ps.deleted=false),'SonarqubeBranchName','STRING','Branch name to be used to send the scanned result on sonarqube project','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube-v1.1.0' and ps."index"=1 and ps.deleted=false),'SonarqubeProjectKey','STRING','Project key of sonarqube account','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube-v1.1.0' and ps."index"=1 and ps.deleted=false),'CheckForSonarAnalysisReport','BOOL','Boolean value - true or false. Set true to poll for generated report from sonarqube.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube-v1.1.0' and ps."index"=1 and ps.deleted=false),'AbortPipelineOnPolicyCheckFailed','BOOL','Boolean value - true or false. Set true to abort on report check failed','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube-v1.1.0' and ps."index"=1 and ps.deleted=false),'UsePropertiesFileFromProject','BOOL','Boolean value - true or false. Set true to use source code sonar-properties file.','t','f',false,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube-v1.1.0' and ps."index"=1 and ps.deleted=false),'SonarqubeEndpoint','STRING','Sonrqube endpoint URL','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube-v1.1.0' and ps."index"=1 and ps.deleted=false),'CheckoutPath','STRING','Checkout path of git material','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube-v1.1.0' and ps."index"=1 and ps.deleted=false),'SonarqubeApiKey','STRING','api key of sonarqube account','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube-v1.1.0' and ps."index"=1 and ps.deleted=false),'TotalSonarqubeIssues','STRING','Total issues in the scanned code result from the sum of vulnerabilities and high hotspots','t','f',false,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube-v1.1.0' and ps."index"=1 and ps.deleted=false),'SonarqubeHighHotspots','STRING','Total number of SonarQube hotspots (HIGH) in the source code','t','f',false,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube-v1.1.0' and ps."index"=1 and ps.deleted=false),'SonarqubeProjectStatus','STRING','Quality gate status of Sonarqube Project ,it may be "ERROR","OK" ,"NONE"','t','f',false,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube-v1.1.0' and ps."index"=1 and ps.deleted=false),'SonarqubeVulnerabilities','STRING','Total number of SonarQube vulnerabilities in the source code','t','f',false,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);


INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube-v1.1.0' and ps."index"=1 and ps.deleted=false), 'GIT_MATERIAL_REQUEST','STRING','git material data',false,true,3,'INPUT','GLOBAL',1 ,'GIT_MATERIAL_REQUEST','f','now()', 1, 'now()', 1);



INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)VALUES (nextval('id_seq_plugin_stage_mapping'),

(SELECT id from plugin_metadata where name='Sonarqube-v1.1.0'), 0,'now()',1,'now()',1);