#!/bin/bash

set -euo pipefail
source $(dirname $0)/../../../utils.sh
ROOT=` realpath $(dirname $0)/../../../../`

cleanup() {
    log "Cleaning up..."
    kubefaas spec destroy
    rm -rf func
    popd
}

if [ -z "${TEST_NOCLEANUP:-}" ]; then
    trap cleanup EXIT
else
    log "TEST_NOCLEANUP is set; not cleaning up test artifacts afterwards."
fi

pushd $(dirname $0)

[ -d specs ]
[ -f specs/README ]
[ -f specs/kubefaas-deployment-config.yaml ]

mkdir -p func
cp $ROOT/examples/nodejs/hello.js func/deploy.js
cp $ROOT/examples/nodejs/hello.js func/source.js

kubefaas spec destroy || true

log "Apply specs"
kubefaas --verbosity 2 spec apply

log "verify deployarchive function works"
kubefaas fn test --name deployarchive

timeout 60s bash -c "waitBuild sourcearchive"

log "verify sourcearchive function works"
kubefaas fn test --name sourcearchive

log "Test PASSED"
