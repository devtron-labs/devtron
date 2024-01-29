UPDATE plugin_pipeline_script SET script=E'PathToCodeDir=/devtroncd$CheckoutPath
cd $PathToCodeDir
if [[ -z "$UsePropertiesFileFromProject" || $UsePropertiesFileFromProject == false ]]
then
  echo "sonar.projectKey=$SonarqubeProjectKey" > sonar-project.properties
fi
docker run \\
--rm \\
-e SONAR_HOST_URL=$SonarqubeEndpoint \\
-e SONAR_LOGIN=$SonarqubeApiKey \\
-v "/$PWD:/usr/src" \\
sonarsource/sonar-scanner-cli

if [[ $CheckForSonarAnalysisReport == true && ! -z "$CheckForSonarAnalysisReport" ]]
then
 status=$(curl -u ${SonarqubeApiKey}:  -sS ${SonarqubeEndpoint}/api/qualitygates/project_status?projectKey=${SonarqubeProjectKey}&branch=master)
 project_status=$(echo $status | jq -r  ".projectStatus.status")
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
fi' WHERE id=(SELECT id FROM plugin_metadata WHERE name='Sonarqube');
DELETE FROM plugin_step_variable WHERE id =(select id from plugin_step_variable where name='SonarContainerImage' and plugin_step_id=(SELECT id FROM plugin_metadata WHERE name='Sonarqube'));






UPDATE plugin_pipeline_script SET script=E'#!/bin/sh
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
    fi' WHERE id=(select script_id FROM plugin_step WHERE plugin_id=(SELECT id FROM plugin_metadata WHERE name='Sonarqube v1.1.0'));
DELETE FROM plugin_step_variable WHERE id =(select id from plugin_step_variable where name='SonarContainerImage' and plugin_step_id=(SELECT id FROM plugin_metadata WHERE name='Sonarqube v1.1.0'));


