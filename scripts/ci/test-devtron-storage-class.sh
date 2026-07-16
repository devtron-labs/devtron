#!/bin/sh

set -eu

ROOT=$(git rev-parse --show-toplevel)
CHART="$ROOT/charts/devtron"
WORK_DIR=$(mktemp -d)
trap 'rm -rf "$WORK_DIR"' 0 1 2 15

STAGED_CHART="$WORK_DIR/devtron"
RENDERED="$WORK_DIR/rendered.yaml"
mkdir -p "$STAGED_CHART"
cp -R "$CHART/." "$STAGED_CHART/"

# First-party PVC templates do not need the chart's remote dependencies. Keep
# this regression test deterministic and runnable without network access.
awk '
    /^dependencies:[[:space:]]*$/ { skip = 1; next }
    skip && /^[A-Za-z][A-Za-z0-9_-]*:[[:space:]]*/ { skip = 0 }
    !skip { print }
' "$STAGED_CHART/Chart.yaml" >"$STAGED_CHART/Chart.yaml.local"
mv "$STAGED_CHART/Chart.yaml.local" "$STAGED_CHART/Chart.yaml"
rm -rf "$STAGED_CHART/charts" "$STAGED_CHART/Chart.lock"

helm template storage-test "$STAGED_CHART" \
    --namespace devtroncd \
    --set-string global.storageClass=ci-storage \
    --set 'installer.modules={cicd}' \
    --set minio.enabled=true \
    --set monitoring.grafana.enabled=true \
    --set devtronEnterprise.enabled=true \
    --set devtronEnterprise.finops.enabled=true >"$RENDERED"

storage_class_counts=$(awk '
    $1 == "storageClassName:" { total++ }
    $1 == "storageClassName:" && $2 == "\"ci-storage\"" { matching++ }
    END { print total + 0, matching + 0 }
' "$RENDERED")

if test "$storage_class_counts" != "6 6"; then
    printf 'expected explicit StorageClass on all 6 Devtron PVCs, found: %s\n' \
        "$storage_class_counts" >&2
    exit 1
fi

printf 'explicit StorageClass rendered on all Devtron PVCs\n'
