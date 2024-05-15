-- Plugin metadata
INSERT INTO "plugin_metadata" ("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_metadata'), 'Devtron SLO Rollback Plugin v1.0.0','Triggers a rollback when a certain criteria is met, when calculated gap value(SLI-SLO) from slo-generator is less than gap value provided by user, then the system does a rollback','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/devtron-logo-plugin.png','f', 'now()', 1, 'now()', 1);

-- Plugin Stage Mapping

INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE name='Devtron SLO Rollback Plugin v1.0.0'),0,'now()', 1, 'now()', 1);

-- Plugin Script

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
    nextval('id_seq_plugin_pipeline_script'),
    E'#!/bin/bash
touch /devtroncd/process/stage-0_out.env
echo "SLO plugin initiating"
timeInMin=$(echo "scale=2; $TIME_WINDOW / 60" | bc)
echo "Sleeping for $timeInMin minutes, slo-report will be ready after $timeInMin minutes"
sleep "$TIME_WINDOW"
pip3 install slo-generator[prometheus]
pip3 install slo-generator[cloud_monitoring]
pip3 install slo-generator[dynatrace]
pip3 install slo-generator[datadog]

SLO_CONFIG_CONTENT=$(cat <<EOF
apiVersion: sre.google.com/v2
kind: ServiceLevelObjective
metadata:
  name: $SLO_NAME
spec:
  description: SLO configuration for prometheus, ratio of number of status codes other than 200 by all http request count
  backend: prometheus
  method: query_sli
  exporters:
  - prometheus
  service_level_indicator:
    expression: >
      (
        sum(
        rate(
          $PROM_METRIC{service="$SERVICE_NAME",status=~\'^2..\'}[window]
        )
      )+
      sum(
        rate(
          $PROM_METRIC{service="$SERVICE_NAME",status=~\'^3..\'}[window]
        )
      )
      )
      /
      sum(
        rate(
          $PROM_METRIC{service="$SERVICE_NAME"}[window]
        )
      )
  goal: $SLO_GOAL
EOF
)
echo "$SLO_CONFIG_CONTENT" > "slo_config.yaml"

SLO_SHARED_CONFIG_CONTENT=$(cat <<EOF
backends:
  prometheus:
    url: $PROM_URL
error_budget_policies:
  default:
    steps:
    - name: $BUDGET_POLICY_NAME
      burn_rate_threshold: $BUDGET_POLICY_THRESHOLD
      alert: $IS_ALERT_ON
      message_alert: Page to defend the SLO
      message_ok: Last hour on track
      window: $TIME_WINDOW
EOF
)
echo "$SLO_SHARED_CONFIG_CONTENT" > "shared_config.yaml"


## rollback script
performRollback() {
    ## first hit api endpoint to get the ciArtifactId for the top artifact https://devtron-11.devtron.info/orchestrator/app/cd-pipeline/61/material/rollback?offset=0&size=20
    ## for any failed resp from any api we will exit
    appId=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.appId\')
    cdPipelineId=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.cdPipelineId\')

    materialResponseCurlReq=$(curl -s -H "token:$DEVTRON_TOKEN" "$DEVTRON_ENDPOINT/orchestrator/app/cd-pipeline/"$cdPipelineId"/material/rollback?offset=0&size=20")
    statusCode=$(echo "$materialResponseCurlReq" | jq -r \'.code\')

    if [ "$statusCode" -eq 200 ]; then
        ciArtifactIdToRollback=$(echo "$materialResponseCurlReq" | jq -r \'.result.ci_artifacts[0].id\')
        wfrId=$(echo "$materialResponseCurlReq" | jq -r \'.result.ci_artifacts[0].wfrId\')
        echo "appId: \'"$appId"\'"
        echo "cdPipelineId: \'"$cdPipelineId"\'"
        echo "ciArtifactIdToRollback: \'"$ciArtifactIdToRollback"\'"
        echo "workflowRunnerId: \'"$wfrId"\'"

        rollBackReleaseCurlReq=$(curl -s -H "token:$DEVTRON_TOKEN" "$DEVTRON_ENDPOINT/orchestrator/app/cd-pipeline/trigger" --data-raw \'{"pipelineId": \'"$cdPipelineId"\',"appId": \'"$appId"\',"ciArtifactId": \'"$ciArtifactIdToRollback"\',"cdWorkflowType": "DEPLOY","deploymentWithConfig": "SPECIFIC_TRIGGER_CONFIG","wfrIdForDeploymentWithSpecificTrigger": \'"$wfrId"\'}\')
        statusCode=$(echo "$rollBackReleaseCurlReq" | jq -r \'.code\')

        if [ "$statusCode" -ne 200 ]; then
            errorUserMessage=$(echo "$rollBackReleaseCurlReq" | jq -r \'.errors[0].userMessage\')
            echo "rollBack failed with the following error: "$errorUserMessage""
            exit 1
        else
            echo "rollBack executed successfully"
            return
        fi
    else
        errorMsg=$(echo "$materialResponseCurlReq" | jq -r \'.errors[0].userMessage\')
        echo "fetching latest ci artifact failed with error: "$errorMsg""
        echo "exiting"
        exit 1
    fi
}


attempt=1
while [ "$attempt" -le "$ATTEMPT_COUNT" ]; do

sloGeneratedOutput=$(slo-generator compute -f ./slo_config.yaml -c ./shared_config.yaml)

echo "slo-generator output:"
slo-generator compute -f ./slo_config.yaml -c ./shared_config.yaml

# Count the occurrences of "ERROR" in the output
error_count=$(echo "$sloGeneratedOutput" | grep -o \'ERROR\' | wc -l)
if [[ "$error_count" -ge 2 ]]; then
    echo "SLO report generation failed. Multiple errors detected."
    echo "hint: if error message from generator is Backend returned NO_DATA, then you can check if TIME_WINDOW value is valid or not, try changing TIME_WINDOW value."
    exit 1
fi

# Check if the sloGeneratedOutput contains "INFO" in the second line, if yes then grep the sloGeneratedOutput values to enforce success/failure conditions as provided by the user
if [[ $(echo "$sloGeneratedOutput" | awk \'NR==2 {print $1}\') == "INFO" ]]; then
    # Use awk to extract the values inside the "INFO" part
    info_values=$(echo "$sloGeneratedOutput" | awk \'NR==2 {for (i=2; i<=NF; i++) print $i}\')
    echo "$info_values"> info.txt
    ##gap value from slo generator
    gap=$(grep -n \'\' info.txt  | grep \'^15:\' | cut -d \':\' -f 2-)
    # Remove percent sign if it exists
    gap=${gap%\%}
    sign=$(echo "$gap" | grep -o \'[+-]\')
    ## user provided GAP value, can be negative or positive, positive can be with a plus sign or without plus sign so appending + to that value for comparison
    ## operating on gap provided by user
    gapUser=$THRESHOLD_GAP_VALUE
    gapUserLength=$(echo "${#gapUser}")
    if [[ "$gapUserLength" -eq 0 ]]; then
        echo "exiting as THRESHOLD_GAP_VALUE not provided"
        exit 1
    fi
    signUser=$(echo "$gapUser" | grep -o \'[+-]\')
    signUserLength=$(echo "${#signUser}")
    if [[ "$signUserLength" -eq 0 ]]; then
        gapUser="+"$gapUser
    fi
    echo "calculated GAP value from slo-generator(SLI-SLO): "$gap""
    echo "user defined GAP value: "$gapUser""
    # we will do rollback in last attempt
    if [[ "$attempt" -eq "$ATTEMPT_COUNT" ]]; then
        # assuming used defined GAP value will be a positive number, currently for negative user defined GAP values this logic will misbhave.

        floatingDecimalGap=$(echo "${gap:1}" | awk \'{printf "%.2f", $0}\')
        floatingDecimalUserGap=$(echo "${gapUser:1}" | awk \'{printf "%.2f", $0}\')

        echo "decimal of calculated GAP value from slo-generator(SLI-SLO): "$floatingDecimalGap""
        echo "decimal of user defined GAP value: "$floatingDecimalUserGap""

        if awk \'BEGIN { exit !(\'"$floatingDecimalGap"\' < \'"$floatingDecimalUserGap"\') }\'; then
            # we will do rollback ireespective of sign, since the integral value is less than user defined GAP value
            echo "============= Performing rollback as the condition has been met ============="
            performRollback
            exit 0
        else
            # when comparing the calculated GAP (e.g., -9) and the user GAP (e.g., +8), even though the absolute values suggest 9 is greater than 8,
            # the sign difference indicates a rollback condition.
            echo "checking sign of calculated gap from slo-generator for possible rollback, sign: $sign"
            if [[ "$sign" == "-" ]]; then
                # rollback condition
                echo "============= Performing rollback as the condition has been met ============="
                performRollback
                exit 0
            fi
        fi
    fi
else
    echo "making another attempt to generate SLO, "$attempt" of "$ATTEMPT_COUNT" attempts left"
fi
attempt=$((attempt + 1))
sleep "$RETRY_DELAY"
done
echo -e "\033[1m======== Maximum retries reached. ========"
echo "no need to execute rollback"
exit 0
',
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);


--Plugin Step

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Devtron SLO Rollback Plugin v1.0.0'),'Step 1','Runnig the plugin','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


-- Input Variables

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format, description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by)VALUES
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron SLO Rollback Plugin v1.0.0' and ps."index"=1 and ps.deleted=false),'DEVTRON_TOKEN','STRING','Enter Devtron API Token with required permissions.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron SLO Rollback Plugin v1.0.0' and ps."index"=1 and ps.deleted=false),'DEVTRON_ENDPOINT','STRING','Enter the URL of Devtron Dashboard for.eg (https://abc.xyz).','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron SLO Rollback Plugin v1.0.0' and ps."index"=1 and ps.deleted=false),'RETRY_DELAY','NUMBER','Enter a retry delay value, system will sleep for RETRY_DELAY seconds before triggering any further attempts.','t','f',2,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron SLO Rollback Plugin v1.0.0' and ps."index"=1 and ps.deleted=false),'THRESHOLD_GAP_VALUE','STRING','Expecting a positive value e.g +5.00 or +6, Enter a threshold gap, this value is compared to the calculated gap value from slo-generator, if gap from slo-generator is less than user defined gap value then rollback is triggered.','t','f','+2.00',null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron SLO Rollback Plugin v1.0.0' and ps."index"=1 and ps.deleted=false),'TIME_WINDOW','NUMBER','This variable is the duration that the post-cd will sleep for before generating slo-report from respective exporter, this is also the duration for which the metrics is desired for.','t','f',300,null,'INPUT','NEW',null,1,null,null, 'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron SLO Rollback Plugin v1.0.0' and ps."index"=1 and ps.deleted=false),'IS_ALERT_ON','STRING','Is the alert on.','t','f','f',null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron SLO Rollback Plugin v1.0.0' and ps."index"=1 and ps.deleted=false),'BUDGET_POLICY_THRESHOLD','NUMBER','The error budget is the total allowable amount of errors or deviations from the desired service level within a specified timeframe.', 't','f',10,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron SLO Rollback Plugin v1.0.0' and ps."index"=1 and ps.deleted=false),'BUDGET_POLICY_NAME','STRING','Budget policy name.', 't','f','budget_policy',null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron SLO Rollback Plugin v1.0.0' and ps."index"=1 and ps.deleted=false),'PROM_URL','STRING','Prometheus endpoint url.', 't','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron SLO Rollback Plugin v1.0.0' and ps."index"=1 and ps.deleted=false),'SLO_GOAL','NUMBER','Set the SLO goal for your system. This is the metric against which your SLI will be measured.(SLO=SLI-GAP)', 't','f',0.9,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron SLO Rollback Plugin v1.0.0' and ps."index"=1 and ps.deleted=false),'SERVICE_NAME','STRING','Name of the service for which you want to apply your filter to scrape data from.', 't','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron SLO Rollback Plugin v1.0.0' and ps."index"=1 and ps.deleted=false),'PROM_METRIC','STRING','Name of prometheus metric to be used.', 't','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron SLO Rollback Plugin v1.0.0' and ps."index"=1 and ps.deleted=false),'SLO_NAME','STRING','Set a Service level objective name.', 't','f','slo_sample',null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron SLO Rollback Plugin v1.0.0' and ps."index"=1 and ps.deleted=false),'ATTEMPT_COUNT','NUMBER','Set an attempt number, if slo-report is generated without any error then it will retry ATTEMPT_COUNT number of times before doing any rollbacks if necessary.', 't','f',2,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

