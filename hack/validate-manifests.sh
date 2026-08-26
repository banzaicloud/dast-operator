#!/usr/bin/env bash
set -euo pipefail

TARGET_K8S_VERSION="${PLUTO_TARGET_K8S_VERSION:-v1.31.0}"
CHART_DIR="charts/dast-operator"

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

helm template dast-operator "$CHART_DIR" \
  --values "$CHART_DIR/ci/test-values.yaml" \
  > "$workdir/chart-templates.yaml"

cp "$CHART_DIR"/crds/*.yaml "$workdir/"

echo "Validating rendered chart against Kubernetes ${TARGET_K8S_VERSION}"
pluto detect-files -d "$workdir" --target-versions "k8s=${TARGET_K8S_VERSION}"
