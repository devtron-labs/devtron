#!/bin/bash

# Test various endpoint combinations to find the correct ones
echo "Testing various endpoint combinations..."

BASE_URL="https://devtron-ent-2.devtron.info"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IkFQSS1UT0tFTjphZG1pbiIsInZlcnNpb24iOiIxIiwiaXNzIjoiYXBpVG9rZW5Jc3N1ZXIiLCJleHAiOjE3NTY1NTc1NzR9.ZHhQdhXpGygCOiO7rDah0mBB7zZYZ3y9WlJL9egRfq4"

# Test the original auth endpoint
echo "1. Testing original auth endpoint:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: */*' \
  -H 'Authorization: Bearer '"$TOKEN" \
  "$BASE_URL/orchestrator/devtron/auth/verify/v2"

# Test without orchestrator prefix
echo -e "\n2. Testing without orchestrator prefix:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: */*' \
  -H 'Authorization: Bearer '"$TOKEN" \
  "$BASE_URL/devtron/auth/verify/v2"

# Test with just the base path
echo -e "\n3. Testing with just base path:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: */*' \
  -H 'Authorization: Bearer '"$TOKEN" \
  "$BASE_URL/auth/verify/v2"

# Test app store endpoint with different paths
echo -e "\n4. Testing app store with orchestrator prefix:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer '"$TOKEN" \
  "$BASE_URL/orchestrator/devtron/app-store/discover"

echo -e "\n5. Testing app store without orchestrator prefix:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer '"$TOKEN" \
  "$BASE_URL/devtron/app-store/discover"

echo -e "\n6. Testing app store with just base path:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer '"$TOKEN" \
  "$BASE_URL/app-store/discover"

# Test a simple health check endpoint
echo -e "\n7. Testing health check endpoint:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: application/json' \
  "$BASE_URL/health"

echo -e "\n8. Testing health check with orchestrator:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: application/json' \
  "$BASE_URL/orchestrator/health"

echo -e "\nEndpoint testing completed. Look for 200 status codes to identify working endpoints." 