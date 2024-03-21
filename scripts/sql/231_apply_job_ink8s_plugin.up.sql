INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Apply JOB in k8s v1.0.0','Apply custom jobs in k8s cluster.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/devtron-logo-plugin.png',false,'now()',1,'now()',1);


INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)VALUES (nextval('id_seq_plugin_stage_mapping'),
(SELECT id from plugin_metadata where name='Apply JOB in k8s v1.0.0'), 0,'now()',1,'now()',1);


INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES ( nextval('id_seq_plugin_pipeline_script'),
E'#!/bin/sh
RUN_MIGRATION=$(echo $CI_CD_EVENT | jq -r \'.commonWorkflowRequest.extraEnvironmentVariables.RUN_MIGRATION\')
echo $RUN_MIGRATION
if [ "$RUN_MIGRATION" == "true" ]; then
    # Configuration variables
    NAMESPACE=$Namespace
    NAME=$JobName
    RUN_COMMAND=$RunCommand
    BUILD_ARC=$BuildArch
    SERVICE_ACCOUNT=$ServiceAccount
    HEALTH_ENDPOINT=$HealthEndpoint
    ENV_PATH=$EnvPath
    JOB_TEMPLATE=$JobTemplatePath
    MAX_ATTEMPTS=$MaxAttempts 
    SLEEP_TIME=$SleepTime

    if [ -z "$NAMESPACE" ];then
        echo "Exiting due to Namespace not specified".
        exit 1
    elif [ -z "$NAME" ];then
        echo "Exiting due to JobName not specified".
        exit 1
    elif [ -z "$RUN_COMMAND" ];then
        echo "Exiting due to RunCommand not specified".
        exit 1
    elif [ -z "$BUILD_ARC" ];then
        echo "Exiting due to BuildArch not specified".
        exit 1
    elif [ -z "$SERVICE_ACCOUNT" ];then
        echo "Exiting due to ServiceAccount not specified".
        exit 1
    elif [ -z "$HEALTH_ENDPOINT" ];then 
        echo "Exiting due to HealthEndpoint not specified".
        exit 1
    elif [ -z "$ENV_PATH" ];then
        echo "Exiting due to EnvPath not specified".
        exit 1
    elif [ -z "$KubeConfig" ];then
        echo "Exiting due to KubeConfig not specified".
        exit 1
    fi

    if [ -z "$MAX_ATTEMPTS" ];then 
        echo "MaxAttempts not specified using the default one i.e. 20" #Will set these values in SQL
    fi
    if [ -z "$SLEEP_TIME" ];then
        echo "SleepTime not specified using the default one i.e. 15"    #Will set these values in SQL
    fi

    echo "Running migration job"

    # Devtron Config
    cd /devtroncd
    touch kubeconfig.yaml
    touch kubeconfig.txt
    echo $KubeConfig > kubeconfig.txt
    cat kubeconfig.txt | base64 -d > kubeconfig.yaml

    # Get the system architecture
    architecture=$(uname -m)

    # Check if the architecture is AMD or ARM
    if [[ $architecture == "x86_64" || $architecture == "amd64" ]]; then
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    elif [[ $architecture == "aarch64" || $architecture == "arm64" ]]; then
         curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/arm64/kubectl"
    else
        echo "Unknown system architecture: $architecture"
        exit 1
    fi
    install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

    # Custom Variables
    export tag=$(echo $CI_CD_EVENT | jq --raw-output .commonWorkflowRequest.dockerImageTag)
    export repo=$(echo $CI_CD_EVENT | jq --raw-output .commonWorkflowRequest.dockerRepository)
    export registry=$(echo $CI_CD_EVENT | jq --raw-output .commonWorkflowRequest.dockerRegistryURL)

    echo $registry/$repo:$tag
    IMAGE_TAG=$registry/$repo:$tag


    if [ $JobTemplateScoped ];then
        echo "Using JOB template from scoped variable"
        touch job-template.yaml
        touch temp.txt
        echo $JobTemplateScoped > temp.txt
        cat temp.txt | base64 -d > job-template.yaml
    else
        if [ $JOB_TEMPLATE ];then
            echo "Using external job template from repo"
            touch job-template.yaml
            echo "Path to jobtemplate: $JOB_TEMPLATE"
            cat $JOB_TEMPLATE > job-template.yaml
        else
            echo "Using internal job template"
            echo "No job template specified. Using the default"
            touch jobtemplate.txt
            touch job-template.yaml
            default_job_template="YXBpVmVyc2lvbjogYmF0Y2gvdjEKa2luZDogSm9iCm1ldGFkYXRhOgogIG5hbWU6IFZBUi1KT0ItTkFNRS1SQU5ET00tU1RSSU5HCiAgbmFtZXNwYWNlOiBWQVItTkFNRVNQQUNFCnNwZWM6CiAgYmFja29mZkxpbWl0OiAwCiAgYWN0aXZlRGVhZGxpbmVTZWNvbmRzOiAxMDgwMAogIHRlbXBsYXRlOgogICAgc3BlYzoKICAgICAgYWZmaW5pdHk6CiAgICAgICAgbm9kZUFmZmluaXR5OgogICAgICAgICAgcmVxdWlyZWREdXJpbmdTY2hlZHVsaW5nSWdub3JlZER1cmluZ0V4ZWN1dGlvbjoKICAgICAgICAgICAgbm9kZVNlbGVjdG9yVGVybXM6CiAgICAgICAgICAgICAgLSBtYXRjaEV4cHJlc3Npb25zOgogICAgICAgICAgICAgICAgICAtIGtleTogbm9kZXR5cGUKICAgICAgICAgICAgICAgICAgICBvcGVyYXRvcjogSW4KICAgICAgICAgICAgICAgICAgICB2YWx1ZXM6CiAgICAgICAgICAgICAgICAgICAgICAtIFZBUi1CVUlMRC1BUkMKICAgICAgY29udGFpbmVyczoKICAgICAgLSBuYW1lOiBWQVItSk9CLU5BTUUKICAgICAgICBpbWFnZTogVkFSLUlNQUdFLVRBRwogICAgICAgIGFyZ3M6CiAgICAgICAgICAtIC9iaW4vc2gKICAgICAgICAgIC0gLWMKICAgICAgICAgIC0gVkFSLU1JR1JBVElPTi1SVU4tQ09NTUFORAogICAgICAgIGVudjoKICAgICAgICAgIC0gbmFtZTogRU5WX1BBVEgKICAgICAgICAgICAgdmFsdWU6IFZBUi1FTlYtUEFUSAogICAgICAgICAgLSBuYW1lOiBWQVVMVF9UT0tFTgogICAgICAgICAgICB2YWx1ZUZyb206CiAgICAgICAgICAgICAgc2VjcmV0S2V5UmVmOgogICAgICAgICAgICAgICAgbmFtZTogdmF1bHQtc2VjcmV0CiAgICAgICAgICAgICAgICBrZXk6IFZBVUxUX1RPS0VOCiAgICAgICAgICAtIG5hbWU6IFZBVUxUX1VSTAogICAgICAgICAgICB2YWx1ZUZyb206CiAgICAgICAgICAgICAgY29uZmlnTWFwS2V5UmVmOgogICAgICAgICAgICAgICAgbmFtZTogdmF1bHQtdXJsCiAgICAgICAgICAgICAgICBrZXk6IFZBVUxUX1VSTAogICAgICAgIHJlc291cmNlczoKICAgICAgICAgIGxpbWl0czoKICAgICAgICAgICAgY3B1OiAiMiIKICAgICAgICAgICAgbWVtb3J5OiA0R2kKICAgICAgICAgIHJlcXVlc3RzOgogICAgICAgICAgICBjcHU6ICIyIgogICAgICAgICAgICBtZW1vcnk6IDRHaQogICAgICByZXN0YXJ0UG9saWN5OiBOZXZlcgogICAgICBzZXJ2aWNlQWNjb3VudE5hbWU6IFZBUi1TRVJWSUNFLUFDQ09VTlQ="
            echo $default_job_template > jobtemplate.txt
            cat jobtemplate.txt | base64 -d > job-template.yaml
        fi

    fi
    
    set +o pipefail
    RANDOM_STRING=$(openssl rand -base64 6 | tr -dc a-z | fold -w 4 | head -n 1)
    set -o pipefail

    JOB_NAME="${NAME}-${RANDOM_STRING}"

    sed -i -e "s|VAR-JOB-NAME-RANDOM-STRING| ${JOB_NAME}|g" \\
        -e "s|VAR-JOB-NAME|${NAME}|g" \\
        -e "s|VAR-NAMESPACE|${NAMESPACE}|g" \\
        -e "s|VAR-BUILD-ARC|${BUILD_ARC}|g" \\
        -e "s|VAR-IMAGE-TAG|${IMAGE_TAG}|g" \\
        -e "s|VAR-ENV-PATH|${ENV_PATH}|g" \\
        -e "s|VAR-SERVICE-ACCOUNT|${SERVICE_ACCOUNT}|g" \\
        -e "s|VAR-MIGRATION-RUN-COMMAND|${RUN_COMMAND}|g" job-template.yaml

    FILE_PATH=job-template.yaml
    cat job-template.yaml

    echo "Applying job YAML..."
    kubectl apply -f "$FILE_PATH" --kubeconfig /devtroncd/kubeconfig.yaml

    echo "Waiting for the pod to be in the Running state..."
    for ATTEMPT in $(seq 1 $MAX_ATTEMPTS); do
        echo "Checking pod status, attempt $ATTEMPT of $MAX_ATTEMPTS..."
        POD_NAME=$(kubectl get pods --kubeconfig /devtroncd/kubeconfig.yaml -n $NAMESPACE --selector=job-name=$JOB_NAME -o jsonpath=\'{.items[0].metadata.name}\')
        if [ -z "$POD_NAME" ]; then
            echo "Pod not found yet. Waiting..."
            sleep $SLEEP_TIME
            continue
        fi

        POD_STATUS=$(kubectl get pod --kubeconfig /devtroncd/kubeconfig.yaml $POD_NAME -n $NAMESPACE -o jsonpath=\'{.status.phase}\')
        if [ "$POD_STATUS" = "Running" ]; then
            echo "Pod $POD_NAME is running."
            POD_RUNNING=1
            break
        else
            echo "Pod $POD_NAME is not ready yet. Status: $POD_STATUS"
            sleep $SLEEP_TIME
        fi
    done

    if [ "$POD_STATUS" != "Running" ]; then
        echo "Pod did not reach running state within the allowed attempts."
        exit 1
    fi

    # Perform health check
    echo "Performing health check..."
# Ensure the loop only starts if the pod is in a Running state
    if [ $POD_RUNNING -eq 1 ]; then
        POD_IPS=$(kubectl get pods --kubeconfig /devtroncd/kubeconfig.yaml -n $NAMESPACE --selector=job-name=$JOB_NAME -o jsonpath=\'{.items[*].status.podIP}\')
        echo "Pod IPs: $POD_IPS"
        
        HEALTHY=0
        echo "Performing health check..."
        for ATTEMPT in $(seq 1 $MAX_ATTEMPTS); do
            FULL_URL="http://$POD_IPS$HEALTH_ENDPOINT"
            echo "Checking URL: $FULL_URL"
            STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$FULL_URL") || true
            echo "Attempt $ATTEMPT: Received status $STATUS"
            
            if [ "$STATUS" = "200" ]; then
                echo "Pod health check PASSED"
                HEALTHY=1
                break
            elif [ "$STATUS" = "000" ]; then
                echo "Attempt $ATTEMPT: Unable to connect to the pod. Retrying..."
            else
                echo "Attempt $ATTEMPT: Waiting for pod to become healthy... Status: $STATUS"
            fi
            sleep $SLEEP_TIME
        done

        if [ $HEALTHY -ne 1 ]; then
            echo "Pod health check FAILED after $MAX_ATTEMPTS attempts"
            kubectl delete job $JOB_NAME -n $NAMESPACE --kubeconfig /devtroncd/kubeconfig.yaml
            exit 1
        fi
    else
        echo "Pod did not reach healthy state within the allowed attempts."
        exit 1
    fi


    echo "Migration completed successfully."
else
    echo "Skipping Migration"
fi
'
,
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);
INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Apply JOB in k8s v1.0.0'),'Step 1','Running the plugin','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);
INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES 
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Apply JOB in k8s v1.0.0' and ps."index"=1 and ps.deleted=false),'Namespace',      'STRING','The namespace where the JOB is to be applied.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Apply JOB in k8s v1.0.0' and ps."index"=1 and ps.deleted=false),'JobName',        'STRING','The name of the JOB to run','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Apply JOB in k8s v1.0.0' and ps."index"=1 and ps.deleted=false),'RunCommand',     'STRING','Run command for the JOB','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Apply JOB in k8s v1.0.0' and ps."index"=1 and ps.deleted=false),'BuildArch',      'STRING','Build architecture.',  't','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Apply JOB in k8s v1.0.0' and ps."index"=1 and ps.deleted=false),'ServiceAccount', 'STRING','Service account.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Apply JOB in k8s v1.0.0' and ps."index"=1 and ps.deleted=false),'HealthEndpoint', 'STRING','Health endpoint for health-check.', 't','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Apply JOB in k8s v1.0.0' and ps."index"=1 and ps.deleted=false),'EnvPath',        'STRING','Path of env variables.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Apply JOB in k8s v1.0.0' and ps."index"=1 and ps.deleted=false),'KubeConfig',     'STRING','base64 encoded KubeConfig of the cluster where the JOB is to be applied.',  't','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Apply JOB in k8s v1.0.0' and ps."index"=1 and ps.deleted=false),'JobTemplateScoped','STRING','base64 encoded job template through scoped variable. Will use default if not provided.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Apply JOB in k8s v1.0.0' and ps."index"=1 and ps.deleted=false),'JobTemplatePath','STRING','Path of the JOB template.Will use default if not provided.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Apply JOB in k8s v1.0.0' and ps."index"=1 and ps.deleted=false),'MaxAttempts',    'NUMBER','Maximum attempts to check the JOB status.','t','t',20,  null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Apply JOB in k8s v1.0.0' and ps."index"=1 and ps.deleted=false),'SleepTime',      'NUMBER','Time interval between each health check.','t','t',15,  null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);