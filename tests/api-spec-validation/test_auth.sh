#!/bin/bash

# Test authentication against live Devtron server
echo "Testing authentication against live Devtron server..."

# Test the auth endpoint from the curl example
curl -s -o /dev/null -w "%{http_code}" \
  -H 'accept: */*' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IkFQSS1UT0tFTjphZG1pbiIsInZlcnNpb24iOiIxIiwiaXNzIjoiYXBpVG9rZW5Jc3N1ZXIiLCJleHAiOjE3NTY1NTc1NzR9.ZHhQdhXpGygCOiO7rDah0mBB7zZYZ3y9WlJL9egRfq4' \
  'https://devtron-ent-2.devtron.info/orchestrator/devtron/auth/verify/v2'

echo " - Auth endpoint test completed"

# Test a simple GET endpoint
echo "Testing app store discover endpoint..."
curl -s -o /dev/null -w "%{http_code}" \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IkFQSS1UT0tFTjphZG1pbiIsInZlcnNpb24iOiIxIiwiaXNzIjoiYXBpVG9rZW5Jc3N1ZXIiLCJleHAiOjE3NTY1NTc1NzR9.ZHhQdhXpGygCOiO7rDah0mBB7zZYZ3y9WlJL9egRfq4' \
  'https://devtron-ent-2.devtron.info/orchestrator/devtron/app-store/discover'

echo " - App store endpoint test completed"

echo "Authentication tests completed. If you see 200 status codes, authentication is working." 