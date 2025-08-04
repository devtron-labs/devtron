#!/bin/bash

# Test using the exact curl command provided by the user
echo "Testing with exact curl command from user..."

# Test the auth endpoint with cookies (like the original curl)
echo "1. Testing auth endpoint with cookies:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: */*' \
  -H 'accept-language: en-US,en;q=0.9,hi;q=0.8' \
  -b '_ga=GA1.1.654831891.1739442610; _ga_5WWMF8TQVE=GS1.1.1742452726.1.1.1742452747.0.0.0; _ga_1LZ3NN5V05=GS1.1.1742452728.1.1.1742452747.0.0.0; _ga_SRRDVHWN2V=GS1.1.1744116562.18.0.1744116562.0.0.0; _ga_GZL0X1L6N8=GS2.1.s1746514541$o4$g1$t1746514550$j0$l0$h0; argocd.token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTM1Mjc5OTMsImp0aSI6ImI2NmNiMjNlLThiMGEtNDJhMS1iMzY0LWE2MGY4YjdiYmY0NCIsImlhdCI6MTc1MzQ0MTU5MywiaXNzIjoiYXJnb2NkIiwibmJmIjoxNzUzNDQxNTkzLCJzdWIiOiJhZG1pbiJ9.JGFD3O-n9NEokWppDeNbsuu9ojjviJojhgRr9qRmXq8' \
  -H 'priority: u=1, i' \
  -H 'referer: https://devtron-ent-2.devtron.info/dashboard/app/list/d' \
  -H 'sec-ch-ua: "Not)A;Brand";v="8", "Chromium";v="138", "Google Chrome";v="138"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "macOS"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: same-origin' \
  -H 'user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36' \
  'https://devtron-ent-2.devtron.info/orchestrator/devtron/auth/verify/v2'

echo -e "\n2. Testing with API token in Authorization header:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: */*' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IkFQSS1UT0tFTjphZG1pbiIsInZlcnNpb24iOiIxIiwiaXNzIjoiYXBpVG9rZW5Jc3N1ZXIiLCJleHAiOjE3NTY1NTc1NzR9.ZHhQdhXpGygCOiO7rDah0mBB7zZYZ3y9WlJL9egRfq4' \
  'https://devtron-ent-2.devtron.info/orchestrator/devtron/auth/verify/v2'

echo -e "\n3. Testing with both cookies and API token:"
curl -s -w "Status: %{http_code}\n" \
  -H 'accept: */*' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6IkFQSS1UT0tFTjphZG1pbiIsInZlcnNpb24iOiIxIiwiaXNzIjoiYXBpVG9rZW5Jc3N1ZXIiLCJleHAiOjE3NTY1NTc1NzR9.ZHhQdhXpGygCOiO7rDah0mBB7zZYZ3y9WlJL9egRfq4' \
  -b '_ga=GA1.1.654831891.1739442610; _ga_5WWMF8TQVE=GS1.1.1742452726.1.1.1742452747.0.0.0; _ga_1LZ3NN5V05=GS1.1.1742452728.1.1.1742452747.0.0.0; _ga_SRRDVHWN2V=GS1.1.1744116562.18.0.1744116562.0.0.0; _ga_GZL0X1L6N8=GS2.1.s1746514541$o4$g1$t1746514550$j0$l0$h0; argocd.token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTM1Mjc5OTMsImp0aSI6ImI2NmNiMjNlLThiMGEtNDJhMS1iMzY0LWE2MGY4YjdiYmY0NCIsImlhdCI6MTc1MzQ0MTU5MywiaXNzIjoiYXJnb2NkIiwibmJmIjoxNzUzNDQxNTkzLCJzdWIiOiJhZG1pbiJ9.JGFD3O-n9NEokWppDeNbsuu9ojjviJojhgRr9qRmXq8' \
  'https://devtron-ent-2.devtron.info/orchestrator/devtron/auth/verify/v2'

echo -e "\nTesting completed. Check the status codes and response bodies." 