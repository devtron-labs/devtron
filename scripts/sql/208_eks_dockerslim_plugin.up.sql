INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'EKS Create Cluster','Plugin to provision a EKS cluster in AWS','PRESET','null',false,'now()',1,'now()',1);

INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_stage_mapping'),(SELECT id from plugin_metadata where name='EKS Create Cluster'), 0,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
     nextval('id_seq_plugin_pipeline_script'),
        $$#!/bin/bash

set -e

CLUSTER_NAME="${ClusterName:-devtron-plugin-cluster}"
VERSION="${Version:-1.25}"
REGION="${Region:-us-west-2}"
ZONES="${Zones:-us-west-2a,us-west-2b,us-west-2c}"
NODEGROUP_NAME="${NodeGroupName:-linux-nodes}"
NODE_TYPE="${NodeType:-m5.large}"
DESIRED_NODES="${DesiredNodes:-1}"
MIN_NODES="${MinNodes:-1}"
MAX_NODES="${MaxNodes:-4}"
USE_IAM_NODE_ROLE=$(echo "$UseIAMNodeRole" | tr "[:upper:]" "[:lower:]")
USE_CONFIG_FILE=$(echo "$UseConfigFile" | tr "[:upper:]" "[:lower:]")
CONFIG_FILE_PATH="${ConfigFilePath}"
AWS_ACCESS_KEY_ID=$AWSAccessKeyId
AWS_SECRET_ACCESS_KEY=$AWSSecretAccessKey

curl --silent --location "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
mv /tmp/eksctl /usr/local/bin

# Check if IAM node role is used
if [ "$USE_IAM_NODE_ROLE" == "true" ]; then
  echo "Using IAM node role for AWS credentials"
  AWS_CLI_CONFIG="/home/tekton/.aws"
  mkdir -p "$AWS_CLI_CONFIG"
else
  # Check if AWS credentials are provided
  if [ -z "$AWSAccessKeyId" ] || [ -z "$AWSSecretAccessKey" ]; then
    echo "Error: AWS credentials not provided. Set USE_IAM_NODE_ROLE=true or provide AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY."
    exit 1
  fi

  # AWS CLI configuration
  # Write AWS credentials to AWS CLI configuration
  echo "exporting aws credentials"
  export AWS_ACCESS_KEY_ID="$AWSAccessKeyId"
  export AWS_SECRET_ACCESS_KEY="$AWSSecretAccessKey"
fi

# Check if using EKS config file
if [ "$USE_CONFIG_FILE" == "true" ]; then
  if [ -z "$CONFIG_FILE_PATH" ]; then
    echo "Error: EKS config file path not provided. Set CONFIG_FILE_PATH when USE_CONFIG_FILE=true."
    exit 1
  fi

  # Create EKS cluster using config file
  echo "************ Using Eksctl config file to create the cluster ***************"
  eksctl create cluster --config-file "/devtroncd/$CONFIG_FILE_PATH" --kubeconfig /devtroncd/kubeconfig

else
  echo "************** Creating Eksctl cluster using the parameters provided in plugin **************"
  # Create EKS cluster using specified parameters
  eksctl create cluster \\
    --name "$CLUSTER_NAME" \\
    --version "$VERSION" \\
    --region "$REGION" \\
    --zones "$ZONES" \\
    --nodegroup-name "$NODEGROUP_NAME" \\
    --node-type "$NODE_TYPE" \\
    --nodes "$DESIRED_NODES" \\
    --nodes-min "$MIN_NODES" \\
    --nodes-max "$MAX_NODES" \\
    --kubeconfig /devtroncd/kubeconfig
fi

# Check if the cluster creation was successful
if [ $? -eq 0 ]; then
  echo "***** Successfully created EKS cluster: $CLUSTER_NAME *****"
  
  # Write kubeconfig to the specified workspace
  
  echo "***** Kubeconfig file written to: /devtroncd/kubeconfig *****"

  echo "********* EKS kubeconfig ********* "
  cat /devtroncd/kubeconfig/kubeconfig.yaml
else
  echo "Error: Failed to create EKS cluster: $CLUSTER_NAME"
  exit 1
fi$$,
        'SHELL',
        'f',
        'now()',
        1,
        'now()',
        1
);


INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='EKS Create Cluster'),'Step 1','Step 1 - EKS Create Cluster','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster' and ps."index"=1 and ps.deleted=false),'ClusterName','STRING','Provide the Cluster Name','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster' and ps."index"=1 and ps.deleted=false),'Version','STRING','Version of the EKS Cluster to create','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster' and ps."index"=1 and ps.deleted=false),'Region','STRING','AWS Region for EKS Cluster','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster' and ps."index"=1 and ps.deleted=false),'Zones','STRING','Availability Zone for EKS Cluster','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster' and ps."index"=1 and ps.deleted=false),'NodeGroupName','STRING','NodeGroup Name for EKS Cluster','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster' and ps."index"=1 and ps.deleted=false),'NodeType','STRING','EC2 instance type for NodeGroup','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster' and ps."index"=1 and ps.deleted=false),'DesiredNodes','STRING','No. of Desired nodes in NodeGroup','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster' and ps."index"=1 and ps.deleted=false),'MinNodes','STRING','No. of Minimum nodes in NodeGroup','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster' and ps."index"=1 and ps.deleted=false),'MaxNodes','STRING','No. of Maximum nodes in NodeGroup','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster' and ps."index"=1 and ps.deleted=false),'ConfigFilePath','STRING','Path for EKS config file','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster' and ps."index"=1 and ps.deleted=false),'UseIAMNodeRole','BOOL','True or False to use IAM Node Role for EKS Cluster creation ','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster' and ps."index"=1 and ps.deleted=false),'UseConfigFile','BOOL','True or False to use ConfigFile for EKS Cluster creation','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster' and ps."index"=1 and ps.deleted=false),'AWSAccessKeyId','STRING','AWS Access Key ID','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster' and ps."index"=1 and ps.deleted=false),'AWSSecretAccessKey','STRING','AWS Secret Access KEY','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);


INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'DockerSlim','This plugin is used to Slim the docker images (Currently this plugin can be used only for docker images not for docker buildx images).','PRESET','null',false,'now()',1,'now()',1);

INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_stage_mapping'),(SELECT id from plugin_metadata where name='DockerSlim'), 0,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
     nextval('id_seq_plugin_pipeline_script'),
        $$#!/bin/bash
httpProbe=$(echo "$HTTPProbe" | tr "[:upper:]" "[:lower:]")
includeFilePath=$IncludePathFile

apk add jq

export tag=$(echo $CI_CD_EVENT | jq --raw-output .commonWorkflowRequest.dockerImageTag)
export repo=$(echo $CI_CD_EVENT | jq --raw-output .commonWorkflowRequest.dockerRepository)

cd /devtroncd

docker pull dslim/slim

if [ "$httpProbe" == "true" ]; then
    if [ -n "$includeFilePath" ]; then
        docker run -i --rm -v /var/run/docker.sock:/var/run/docker.sock -v $PWD:$PWD dslim/slim build --http-probe=true --target $repo:$tag --tag $repo:$tag --continue-after=2 --include-path-file $includeFilePath
    else
        docker run -i --rm -v /var/run/docker.sock:/var/run/docker.sock -v $PWD:$PWD dslim/slim build --http-probe=true --target $repo:$tag --tag $repo:$tag --continue-after=2
    fi
elif [ -n "$includeFilePath" ]; then
    docker run -i --rm -v /var/run/docker.sock:/var/run/docker.sock -v $PWD:$PWD dslim/slim build --http-probe=false --target $repo:$tag --tag $repo:$tag --continue-after=2 --include-path-file $includeFilePath
else
    docker run -i --rm -v /var/run/docker.sock:/var/run/docker.sock -v $PWD:$PWD dslim/slim build --http-probe=false --target $repo:$tag --tag $repo:$tag --continue-after=2
fi

echo "Docker-slim images built"$$,
        'SHELL',
        'f',
        'now()',
        1,
        'now()',
        1
);


INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='DockerSlim'),'Step 1','Step 1 - DockerSlim','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='DockerSlim' and ps."index"=1 and ps.deleted=false),'HTTPProbe','BOOL','Is port expose or not in Dockerfile','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='DockerSlim' and ps."index"=1 and ps.deleted=false),'IncludePathFile','STRING','File path contains including path for dockerslim build flag --include-path-file','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);
