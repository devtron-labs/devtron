UPDATE plugin_pipeline_script SET script = '#!/bin/sh
echo "IMAGE SCAN"

perform_curl_request() {
    local attempt=1
    while [ "$attempt" -le "$MAX_RETRIES" ]; do
        response=$(curl -s -w "\n%{http_code}" -X POST $IMAGE_SCANNER_ENDPOINT/scanner/image -H "Content-Type: application/json" -d "{\"image\": \"$DEST\", \"imageDigest\": \"$DIGEST\", \"pipelineId\" : $PIPELINE_ID, \"userId\": $TRIGGERED_BY, \"dockerRegistryId\": \"$DOCKER_REGISTRY_ID\" }")
        http_status=$(echo "$response" | tail -n1)
        if [ "$http_status" = "200" ]; then
            echo "Vulnerability Scanning request successful."
            return 0
        else
            echo "Attempt $attempt: Vulnerability Scanning request failed with HTTP status code $http_status"
            echo "Response Body: $response"
            attempt=$((attempt + 1))
            sleep "$RETRY_DELAY"
        fi
    done
    echo -e "\033[1m======== Maximum retries reached. Vulnerability Scanning request failed ========"
    exit 1
}
perform_curl_request'
WHERE id = 13;