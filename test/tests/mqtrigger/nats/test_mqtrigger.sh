#!/bin/bash

#
# Create a function and trigger it using NATS
# 

set -euo pipefail
source $(dirname $0)/../../../utils.sh
set +x

TEST_ID=$(generate_test_id)
echo "TEST_ID = $TEST_ID"

ROOT=$(dirname $0)/../../..
DIR=$(dirname $0)

clusterID="kubefaasMQTrigger"
pubClientID="clientPub-$TEST_ID"
subClientID="clientSub-$TEST_ID"
topic="foo.bar$TEST_ID"
resptopic="foo.foo$TEST_ID"
#KUBEFAAS_NATS_STREAMING_URL="http://defaultKubefaasAuthToken@$(minikube ip):4222"
expectedRespOutput="subject:\"$resptopic\" data:\"Hello, World!\""

env=nodejs-$TEST_ID
fn=hello-$TEST_ID
mqt=mqt-$TEST_ID

cleanup() {
    log "Cleaning up..."
    clean_resource_by_id $TEST_ID
}

if [ -z "${TEST_NOCLEANUP:-}" ]; then
    trap cleanup EXIT
else
    log "TEST_NOCLEANUP is set; not cleaning up test artifacts afterwards."
fi

log "Creating nodejs env"
kubefaas env create --name $env --image $NODE_RUNTIME_IMAGE

log "Creating function"
kubefaas fn create --name $fn --env $env --code $DIR/main.js --method GET

log "Creating message queue trigger"
kubefaas mqtrigger create --name $mqt --function $fn --mqtype "nats-streaming" --topic $topic --resptopic $resptopic

# wait until nats trigger is created
sleep 5

#
# Send a message
#
log "Sending message"
go run $DIR/stan-pub/main.go -s $KUBEFAAS_NATS_STREAMING_URL -c $clusterID -id $pubClientID $topic ""

#
# Wait for message on response topic 
#
log "Waiting for response"
response=$(timeout 30s go run $DIR/stan-sub/main.go --last -s $KUBEFAAS_NATS_STREAMING_URL -c $clusterID -id $subClientID $resptopic 2>&1 || true)
echo "$response"
echo "$response" | grep "$expectedRespOutput"

log "Deleting message queue trigger"
kubefaas mqtrigger delete --name $mqt

log "Test PASSED"
