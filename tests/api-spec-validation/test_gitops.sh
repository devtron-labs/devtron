#!/bin/bash

# Test the gitops endpoint with the new ArgoCD token
echo "Testing gitops endpoint with new ArgoCD token..."

BASE_URL="https://devtron-ent-2.devtron.info"
NEW_ARGO_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTQwNTE5NjMsImp0aSI6ImQwZjU0OGYyLWIzNDItNGUxNy05MzRhLWU0MzY3ZTE2ZTRlZCIsImlhdCI6MTc1Mzk2NTU2MywiaXNzIjoiYXJnb2NkIiwibmJmIjoxNzUzOTY1NTYzLCJzdWIiOiJhZG1pbiJ9.dbLq_5lnKnUKI55bg3dIkcIdLj5hVUKSwfU95Aajm7g"

# Test the gitops endpoint with the exact curl command
echo "1. Testing gitops endpoint with exact curl command:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: */*' \
  -H 'accept-language: en-US,en;q=0.9,hi;q=0.8' \
  -b '_ga=GA1.1.654831891.1739442610; _ga_5WWMF8TQVE=GS1.1.1742452726.1.1.1742452747.0.0.0; _ga_1LZ3NN5V05=GS1.1.1742452728.1.1.1742452747.0.0.0; _ga_SRRDVHWN2V=GS1.1.1744116562.18.0.1744116562.0.0.0; _ga_GZL0X1L6N8=GS2.1.s1746514541$o4$g1$t1746514550$j0$l0$h0; argocd.token='"$NEW_ARGO_TOKEN" \
  -H 'priority: u=1, i' \
  -H 'referer: https://devtron-ent-2.devtron.info/dashboard/global-config/gitops' \
  -H 'sec-ch-ua: "Not)A;Brand";v="8", "Chromium";v="138", "Google Chrome";v="138"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "macOS"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: same-origin' \
  -H 'user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36' \
  "$BASE_URL/orchestrator/gitops/config"

echo -e "\n2. Testing gitops endpoint with just the new ArgoCD token:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: */*' \
  -b 'argocd.token='"$NEW_ARGO_TOKEN" \
  "$BASE_URL/orchestrator/gitops/config"

echo -e "\n3. Testing gitops endpoint with API token:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: */*' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IkFQSS1UT0tFTjphZG1pbiIsInZlcnNpb24iOiIxIiwiaXNzIjoiYXBpVG9rZW5Jc3N1ZXIiLCJleHAiOjE3NTY1NTc1NzR9.ZHhQdhXpGygCOiO7rDah0mBB7zZYZ3y9WlJL9egRfq4' \
  "$BASE_URL/orchestrator/gitops/config"

echo -e "\n4. Testing gitops endpoint with both tokens:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: */*' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IkFQSS1UT0tFTjphZG1pbiIsInZlcnNpb24iOiIxIiwiaXNzIjoiYXBpVG9rZW5Jc3N1ZXIiLCJleHAiOjE3NTY1NTc1NzR9.ZHhQdhXpGygCOiO7rDah0mBB7zZYZ3y9WlJL9egRfq4' \
  -b 'argocd.token='"$NEW_ARGO_TOKEN" \
  "$BASE_URL/orchestrator/gitops/config"

echo -e "\n5. Testing the original auth endpoint with new ArgoCD token:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: */*' \
  -b 'argocd.token='"$NEW_ARGO_TOKEN" \
  "$BASE_URL/orchestrator/devtron/auth/verify/v2"

echo -e "\nTesting completed. Look for 200 status codes to identify working authentication method." 