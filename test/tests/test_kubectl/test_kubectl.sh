#!/bin/bash

set -euo pipefail
source $(dirname $0)/../../utils.sh

TEST_ID=$(generate_test_id)
echo "TEST_ID = $TEST_ID"

tmp_dir="/tmp/test-$TEST_ID"
mkdir -p $tmp_dir

ROOT=$(dirname $0)/../../..

cd $ROOT/test/tests/test_kubectl

cleanup() {
    kubectl delete -f spec-yaml -R || true
}

if [ -z "${TEST_NOCLEANUP:-}" ]; then
    trap cleanup EXIT
else
    log "TEST_NOCLEANUP is set; not cleaning up test artifacts afterwards."
fi

name="go-spec-kubectl"
pkgName="go-b4bbb0e0-2d93-47f0-8c4e-eea644eec2a9"

# cleanup first
cleanup

# apply environment & function
kubectl apply -f spec-yaml -R

# wait for build to finish
timeout 90 bash -c "wait_for_builder $name"
timeout 90 bash -c "waitBuildExpectedStatus $pkgName failed"

cp spec-yaml/function-go.yaml $tmp_dir/function-go.yaml
sed -i 's/gogo/go/g' $tmp_dir/function-go.yaml

kubectl apply -f $tmp_dir/function-go.yaml
timeout 90 bash -c "waitBuild $pkgName"

kubefaas fn test --name $name

log "Test PASSED"
