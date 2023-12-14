INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'GKE Provisioner v1.1.0','Provision a Google Kubernetes Engine cluster within a Google Cloud Platform project. The cluster should be configured with an initial firewall setting designed to permit access only to SSH, ports 80 and 8080, and NodePorts','PRESET','https://raw.githubusercontent.com/ajaydevtron/devtron/main/assets/gke-plugin-icon.png',false,'now()',1,'now()',1);

INSERT INTO plugin_tag (id, name, deleted, created_on, created_by, updated_on, updated_by)
SELECT
    nextval('id_seq_plugin_tag'),
    'Google Kubernetes Engine',
    false,
    'now()',
    1,
    'now()',
    1
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_tag
    WHERE name = 'Google Kubernetes Engine'
);

INSERT INTO plugin_tag (id, name, deleted, created_on, created_by, updated_on, updated_by)
SELECT
    nextval('id_seq_plugin_tag'),
    'GCP',
    false,
    'now()',
    1,
    'now()',
    1
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_tag
    WHERE name = 'GCP'
);

INSERT INTO plugin_tag (id, name, deleted, created_on, created_by, updated_on, updated_by)
SELECT
    nextval('id_seq_plugin_tag'),
    'Kubernetes',
    false,
    'now()',
    1,
    'now()',
    1
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_tag
    WHERE name = 'Kubernetes'
);


INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), (SELECT id FROM plugin_tag WHERE name='Google Kubernetes Engine'), (SELECT id FROM plugin_metadata WHERE name='GKE Provisioner v1.1.0'),'now()', 1, 'now()', 1);
INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), (SELECT id FROM plugin_tag WHERE name='Kubernetes'), (SELECT id FROM plugin_metadata WHERE name='GKE Provisioner v1.1.0'),'now()', 1, 'now()', 1);
INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), (SELECT id FROM plugin_tag WHERE name='GCP'), (SELECT id FROM plugin_metadata WHERE name='GKE Provisioner v1.1.0'),'now()', 1, 'now()', 1);


INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
    nextval('id_seq_plugin_pipeline_script'),
    E'#!/bin/sh
    if [ -z $GcpServiceAccountEncodedCredential ] 
    then
        echo -e "\\n****** The GCP service account has not been provided for provisioning a GKE cluster. Please provide the encoded format of the JSON file for the service account. For instructions on creating a service account and assigning the necessary permissions, refer to the following documentation : https://cloud.google.com/iam/docs/service-accounts-create#iam-service-accounts-create-console"
        exit 0
    fi
    mkdir -p /GoogleCloudPlatform
    echo $GcpServiceAccountEncodedCredential | base64 -d > /GoogleCloudPlatform/serviceaccount.json

    if [ -z $GkeMinNodes ] 
    then
        GkeMinNodes=1
    fi
    if [ -z $GkeMaxNodes ] 
    then
        GkeMaxNodes=3
    fi
    if [ -z $GkeRegion ] 
    then
        GkeRegion="us-central1"
    fi
    if [ -z $GkeMachineType ] 
    then
        GkeMachineType="n1-standard-4"
    fi
    if [ -z $GkeImageType ] 
    then
        GkeImageType="cos"
    fi
    if [ -z $GkeClusterVersion ] 
    then
        GkeClusterVersion="latest"
    fi

    echo -e "\\n********** Provided Parameters to spin up GKE cluster *************"
    echo -e "\\n Project Name: $GcpProjectId\\n Identifier: $Identifier \\n Min Nodes : $GkeMinNodes \\n Max Nodes: $GkeMaxNodes \\n Region: $GkeRegion \\n Machine Type: $GkeMachineType \\n Image type: $GkeImageType \\n Cluster Version: $GkeClusterVersion"
    UNIQUE_STR=$(head /dev/urandom | tr -dc a-z0-9 | head -c 10 ; echo '''')
    export UNIQUE_NAME=$Identifier-$UNIQUE_STR
    echo "Unique name is : $UNIQUE_NAME"

    GkeProvisionCmd="gcloud container clusters create $UNIQUE_NAME --quiet --enable-autoscaling --scopes=cloud-platform  --project=$GcpProjectId  --cluster-version=$GkeClusterVersion --min-nodes=$GkeMinNodes --max-nodes=$GkeMaxNodes --region=$GkeRegion  --machine-type=$GkeMachineType  --image-type=$GkeImageType --num-nodes=1 --network=$UNIQUE_NAME"
    if [ ! -z $GkeNodeServiceAccountName ] 
    then
        GkeProvisionCmd="$GkeProvisionCmd --service-account=$GkeNodeServiceAccountName"
    fi

    echo ''#!/bin/sh'' > /GoogleCloudPlatform/gke_provision.sh
    echo ''gcloud auth activate-service-account --key-file=/GoogleCloudPlatform/serviceaccount.json'' >> /GoogleCloudPlatform/gke_provision.sh
    echo ''
    if [ $? -eq 0 ]; then
    echo "Service account authenticated successfully."
    else
        echo "Service account authentication failed."
        exit
    fi'' >> /GoogleCloudPlatform/gke_provision.sh
    echo "gcloud compute networks create $UNIQUE_NAME --project $GcpProjectId --subnet-mode=auto" >> /GoogleCloudPlatform/gke_provision.sh
    echo "echo ''GKE cluster provision will take at least 5 mins, Please wait ....... ''" >> /GoogleCloudPlatform/gke_provision.sh
    echo $GkeProvisionCmd >> /GoogleCloudPlatform/gke_provision.sh
    echo "gcloud container clusters get-credentials --project=$GcpProjectId --region=$GkeRegion $UNIQUE_NAME" >> /GoogleCloudPlatform/gke_provision.sh
    echo "cat ~/.kube/config > /GoogleCloudPlatform/config" >> /GoogleCloudPlatform/gke_provision.sh
    echo "INSTANCE_TAG=\\$(gcloud compute instances list --project=$GcpProjectId --filter=metadata.cluster-name=$UNIQUE_NAME  --limit=1  --format=get\\(tags.items\\) | tr -d ''\\n'')" >> /GoogleCloudPlatform/gke_provision.sh
    echo "gcloud compute firewall-rules create ports-$UNIQUE_NAME --project=$GcpProjectId --network=$UNIQUE_NAME --allow=tcp:22,tcp:80,tcp:8080,tcp:30000-32767,udp:30000-32767 --target-tags=\\$INSTANCE_TAG" >> /GoogleCloudPlatform/gke_provision.sh
    echo -e "\\n ********** Final script to provision GKE cluster ********** \\n"
    cat /GoogleCloudPlatform/gke_provision.sh
    docker run --rm -v "/GoogleCloudPlatform:/GoogleCloudPlatform" quay.io/devtron/gcloud-util:v1.1.0 sh /GoogleCloudPlatform/gke_provision.sh
    GKE_KUBECONFIG=$(cat /GoogleCloudPlatform/config)
    export GkeKubeconfigFilePath="/GoogleCloudPlatform/config"
    if [ -n "$DisplayGkeKubeConfig" ] && [ "$DisplayGkeKubeConfig" = true ]; 
    then
    echo "********* GKE kubeconfig ********* "
    cat /GoogleCloudPlatform/config
    echo "**********************************"
    fi',
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='GKE Provisioner v1.1.0'),'Step 1','Step 1 - GKE Devtron plugin','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GKE Provisioner v1.1.0' and ps."index"=1 and ps.deleted=false),'GcpServiceAccountEncodedCredential','STRING','GCP service account(base64 encoded) that to be used to create GKE cluster in the project','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GKE Provisioner v1.1.0' and ps."index"=1 and ps.deleted=false),'GkeMinNodes','STRING','The minimum number of nodes in the cluster, default is 1','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GKE Provisioner v1.1.0' and ps."index"=1 and ps.deleted=false),'DisplayGkeKubeConfig','BOOL','Do we want to display the kubeconfig? Value either true or false.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GKE Provisioner v1.1.0' and ps."index"=1 and ps.deleted=false),'Identifier','STRING','A string which identifies the purpose for which this cluster is being created. Used to name other resources created.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GKE Provisioner v1.1.0' and ps."index"=1 and ps.deleted=false),'GkeMaxNodes','STRING','The maximum number of nodes in the cluster, default is 3','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GKE Provisioner v1.1.0' and ps."index"=1 and ps.deleted=false),'GkeNodeServiceAccountName','STRING','The Google Cloud Platform Service Account to be used by the node VMs, If no Service Account is specified, the project default service account is used.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GKE Provisioner v1.1.0' and ps."index"=1 and ps.deleted=false),'GkeRegion','STRING','The region to create the cluster in, default is us-central1 ','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GKE Provisioner v1.1.0' and ps."index"=1 and ps.deleted=false),'GkeMachineType','STRING','The machine type to create, default is n1-standard-4 ','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GKE Provisioner v1.1.0' and ps."index"=1 and ps.deleted=false),'GkeImageType','STRING','The type of image to create the nodes , default is cos','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GKE Provisioner v1.1.0' and ps."index"=1 and ps.deleted=false),'GcpProjectId','STRING','The name of the GCP project in which to create the GKE cluster','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GKE Provisioner v1.1.0' and ps."index"=1 and ps.deleted=false),'GkeClusterVersion','STRING','The GKE k8s version to install, default is latest','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GKE Provisioner v1.1.0' and ps."index"=1 and ps.deleted=false),'GkeKubeconfigFilePath','STRING','The kubeconfig path of gke cluster','t','f',false,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);


INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)VALUES (nextval('id_seq_plugin_stage_mapping'),

(SELECT id from plugin_metadata where name='GKE Provisioner v1.1.0'), 0,'now()',1,'now()',1);

UPDATE plugin_pipeline_script SET script=
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
        sleep 10
        export SonarqubeVulnerabilities=$(curl -u ${SonarqubeApiKey}: --location --request GET "$SonarqubeEndpoint/api/issues/search?componentKeys=$GlobalSonarqubeProjectName&types=VULNERABILITY" | jq ".issues | length")
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
    fi' 
    WHERE id=(select script_id from plugin_step where plugin_id =(select id from plugin_metadata where name='Sonarqube v1.1.0'));