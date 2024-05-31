#
# Copyright (c) 2024. Devtron Inc.
#

kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
