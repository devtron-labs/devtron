INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'EKS Create Cluster v1.0.0','Plugin to provision a EKS cluster in AWS','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/eks-plugin-icon.svg',false,'now()',1,'now()',1);

INSERT INTO plugin_tag (id, name, deleted, created_on, created_by, updated_on, updated_by)
SELECT
    nextval('id_seq_plugin_tag'),
    'AWS EKS',
    false,
    'now()',
    1,
    'now()',
    1
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_tag
    WHERE name = 'AWS EKS'
);

INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") 
VALUES (nextval('id_seq_plugin_tag_relation'), (SELECT id FROM plugin_tag WHERE name='AWS EKS'), (SELECT id FROM plugin_metadata WHERE name='EKS Create Cluster v1.0.0'),'now()', 1, 'now()', 1);

INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_stage_mapping'),(SELECT id from plugin_metadata where name='EKS Create Cluster v1.0.0'), 0,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
     nextval('id_seq_plugin_pipeline_script'),
        $$#!/bin/sh
set -e

ENABLE_PLUGIN=$(echo "$EnablePlugin" | tr "[:upper:]" "[:lower:]")
AUTOMATED_NAME=$(echo "$AutomatedName" | tr "[:upper:]" "[:lower:]")
CLUSTER_NAME="${ClusterName}"
VERSION="${Version}"
REGION="${Region}"
ZONES="${Zones}"
NODEGROUP_NAME="${NodeGroupName:-linux-nodes}"
NODE_TYPE="${NodeType:-m5.large}"
DESIRED_NODES="${DesiredNodes:-1}"
MIN_NODES="${MinNodes:-0}"
MAX_NODES="${MaxNodes:-3}"
USE_IAM_NODE_ROLE=$(echo "$UseIAMNodeRole" | tr "[:upper:]" "[:lower:]")
USE_CONFIG_FILE=$(echo "$UseEKSConfigFile" | tr "[:upper:]" "[:lower:]")
CONFIG_FILE_PATH="${EKSConfigFilePath}"
AWS_ACCESS_KEY_ID=$AWSAccessKeyId
AWS_SECRET_ACCESS_KEY=$AWSSecretAccessKey

if [ "$AUTOMATED_NAME" == "true" ]; then
  if [ -z "$CLUSTER_NAME" ]; then
    echo "Error: CLUSTER_NAME is empty. Exiting the script."
    exit 1
  fi

  # Generate a random suffix for the cluster name
  RANDOM_SUFFIX=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 4)

  # Define the regex pattern
  PATTERN='^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$'

  # Check if the random suffix matches the pattern, if not, regenerate
  while [[ ! "$RANDOM_SUFFIX" =~ $PATTERN ]]; do
    RANDOM_SUFFIX=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 4)
  done

  # Check if the cluster name matches the regex, if not, use a default name
  if [[ ! "$CLUSTER_NAME" =~ $PATTERN ]]; then
    echo "Error: CLUSTER_NAME does not match the required regex. Using a default name."
    CLUSTER_NAME="default-devtron-cluster"
  fi

  CLUSTER_NAME="${CLUSTER_NAME}-${RANDOM_SUFFIX}"
  echo "The random generated cluster name is ${CLUSTER_NAME}"
fi

curl --silent --location "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
mv /tmp/eksctl /usr/local/bin

if [ "$ENABLE_PLUGIN" == "true" ]; then
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
    eksctl create cluster --config-file "/devtroncd/$CONFIG_FILE_PATH" --kubeconfig /devtroncd/kubeconfig.yaml

  else
    if [[ -z "$CLUSTER_NAME" ]]; then
    echo "Error: ClusterName should not be empty. Exiting the script."
    exit 1
    fi
    if [[ -z "$VERSION" ]]; then
    echo "Error: Version should not be empty. Exiting the script."
    exit 1
    fi
    if [[ -z "$REGION" ]]; then
    echo "Error: Region should not be empty. Exiting the script."
    exit 1
    fi
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
      --kubeconfig /devtroncd/kubeconfig.yaml
  fi

  # Check if the cluster creation was successful
  if [ $? -eq 0 ]; then
    echo "***** Successfully created EKS cluster: $CLUSTER_NAME *****"
    export CreatedClusterName=$CLUSTER_NAME
    # Write kubeconfig to the specified workspace
    export EKSKubeConfigPath=/devtroncd/kubeconfig.yaml
  else
    echo "Error: Failed to create EKS cluster: $CLUSTER_NAME"
    exit 1
  fi
else
  echo "Error: Please enable the plugin to create plugin"
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
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='EKS Create Cluster v1.0.0'),'Step 1','Step 1 - EKS Create Cluster','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'EnablePlugin','BOOL','True or False to enable plugin','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'AutomatedName','BOOL','True or False to enabling Random name of the cluster creation based on the ClusterName provided','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'UseIAMNodeRole','BOOL','True or False to use IAM Node Role for EKS Cluster creation ','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'AWSAccessKeyId','STRING','AWS Access Key ID','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'AWSSecretAccessKey','STRING','AWS Secret Access KEY','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'ClusterName','STRING','Provide the Cluster Name','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'Version','STRING','Version of the EKS Cluster to create','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'Region','STRING','AWS Region for EKS Cluster','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'Zones','STRING','Availability Zone for EKS Cluster','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'NodeGroupName','STRING','NodeGroup Name for EKS Cluster','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'NodeType','STRING','EC2 instance type for NodeGroup','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'DesiredNodes','STRING','No. of Desired nodes in NodeGroup','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'MinNodes','STRING','No. of Minimum nodes in NodeGroup','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'MaxNodes','STRING','No. of Maximum nodes in NodeGroup','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'UseEKSConfigFile','BOOL','True or False to use ConfigFile for EKS Cluster creation','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'EKSConfigFilePath','STRING','Path for EKS config file','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'CreatedClusterName','STRING','The EKS cluster created name','t','f',false,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='EKS Create Cluster v1.0.0' and ps."index"=1 and ps.deleted=false),'EKSKubeConfigPath','STRING','The Kubeconfig path of EKS','t','f',false,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);