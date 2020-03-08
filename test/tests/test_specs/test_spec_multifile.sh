#!/bin/bash

set -euo pipefail
source $(dirname $0)/../../utils.sh

TEST_ID=$(generate_test_id)
echo "TEST_ID = $TEST_ID"

tmp_dir="/tmp/test-$TEST_ID"
mkdir -p $tmp_dir

ROOT=$(dirname $0)/../../..

env=python-$TEST_ID
fn=spec-$TEST_ID

cleanup() {
    log "Cleaning up..."
    clean_resource_by_id $TEST_ID
    rm -rf $tmp_dir
}

if [ -z "${TEST_NOCLEANUP:-}" ]; then
    trap cleanup EXIT
else
    log "TEST_NOCLEANUP is set; not cleaning up test artifacts afterwards."
fi

cp -r $ROOT/examples/python/multifile $tmp_dir/
pushd $tmp_dir

kubefaas spec init

log "Creating environment spec"
kubefaas env create --spec --name $env --image $PYTHON_RUNTIME_IMAGE --builder $PYTHON_BUILDER_IMAGE

log "Creating function spec"
kubefaas fn create --spec --name $fn --env $env --deploy "multifile/*" --entrypoint main.main

log "Applying specs"
kubefaas spec apply

log "Checking function's existance"
kubefaas fn list | grep $fn

log "Testing function"
kubefaas fn test --name $fn | grep -i hello

log "Destroying spec objects"
kubefaas spec destroy
popd

log "Test PASSED"
