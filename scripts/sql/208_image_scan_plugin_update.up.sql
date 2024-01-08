INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
        (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Vulnerability Scanning' and ps."index"=1 and ps.deleted=false), 'IMAGE_SCAN_MAX_RETRIES','STRING','image scan max retry count',false,true,'INPUT','GLOBAL',1 ,'IMAGE_SCAN_MAX_RETRIES','f','now()', 1, 'now()', 1),
        (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Vulnerability Scanning' and ps."index"=1 and ps.deleted=false), 'IMAGE_SCAN_RETRY_DELAY','STRING','image scan retry delay (in seconds)',false,true,'INPUT','GLOBAL',1 ,'IMAGE_SCAN_RETRY_DELAY','f','now()', 1, 'now()', 1);

UPDATE plugin_pipeline_script SET script = '#!/bin/sh
echo "IMAGE SCAN"

perform_curl_request() {
    local attempt=1
    while [ "$attempt" -le "$IMAGE_SCAN_MAX_RETRIES" ]; do
        response=$(curl -s -w "\n%{http_code}" -X POST $IMAGE_SCANNER_ENDPOINT/scanner/image -H "Content-Type: application/json" -d "{\"image\": \"$DEST\", \"imageDigest\": \"$DIGEST\", \"pipelineId\" : $PIPELINE_ID, \"userId\": $TRIGGERED_BY, \"dockerRegistryId\": \"$DOCKER_REGISTRY_ID\" }")
        http_status=$(echo "$response" | tail -n1)
        if [ "$http_status" = "200" ]; then
            echo "Vulnerability Scanning request successful."
            return 0
        else
            echo "Attempt $attempt: Vulnerability Scanning request failed with HTTP status code $http_status"
            echo "Response Body: $response"
            attempt=$((attempt + 1))
            sleep "$IMAGE_SCAN_RETRY_DELAY"
        fi
    done
    echo -e "\033[1m======== Maximum retries reached. Vulnerability Scanning request failed ========"
    exit 1
}
perform_curl_request'
WHERE id = 13;