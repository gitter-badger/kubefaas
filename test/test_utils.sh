#!/bin/bash

#
# Test runner. Shell scripts that build kubefaas CLI and server, push a
# docker image to GCR, deploy it on a cluster, and run tests against
# that deployment.
#

set -euo pipefail

# ${BASH_SOURCE[0]} returns the relative path of this file.
ROOT=$(realpath `dirname "${BASH_SOURCE[0]}"`/..)

log_start() {
    echo -e `date +%Y/%m/%d:%H:%M:%S`" start: \r\033[33;1m$2\033[0m"
}

log_end() {
    echo -e `date +%Y/%m/%d:%H:%M:%S`" end:$1\r"
}

getVersion() {
    echo $(git rev-parse HEAD)
}

getDate() {
    echo $(date -u +'%Y-%m-%dT%H:%M:%SZ')
}

getGitCommit() {
    echo $(git rev-parse HEAD)
}

set_ci_build_and_deploy_env() {
    export REPO=kubefaas
    export IMAGE=bundle
    export FETCHER_IMAGE=$REPO/fetcher
    export BUILDER_IMAGE=$REPO/builder
    export PRE_UPGRADE_CHECK_IMAGE=$REPO/pre-upgrade-checks

    export BUILD_ID=${CI_BUILD_NUMBER}
    export TAG=test-${BUILD_ID}
    export PRUNE_INTERVAL=1 # this variable controls the interval to run archivePruner. The unit is in minutes.

    export IMAGE_PULL_POLICY="IfNotPresent"

    export ROUTER_SERVICE_TYPE=NodePort
    export SERVICE_TYPE=NodePort
    export NODE_IP="192.168.1.15" # no need it if LB is supported
    export LB_SUPPORT=false

    export KUBEFAAS_NAMESPACE="kubefaas-${BUILD_ID}"
    export FUNCTION_NAMESPACE="kubefaas-${BUILD_ID}-func"
    export KUBEFAAS_BUILDER_NAMESPACE="kubefaas-builder"

    export CONTROLLER_ADDRESS=
    export CONTROLLER_NODE_PORT=31234
    export ROUTER_NODE_PORT=31235

    TEST_BIN=/tmp/${BUILD_ID}/bin
    mkdir -p ${TEST_BIN}
    export PATH=${TEST_BIN}:${PATH}
}

set_ci_test_env() {
    # kubefaas env
    export KUBEFAAS_URL=http://$(kubectl -n ${KUBEFAAS_NAMESPACE} get svc controller -o jsonpath='{...ip}')
    export KUBEFAAS_ROUTER=$(kubectl -n ${KUBEFAAS_NAMESPACE} get svc router -o jsonpath='{...ip}')
    export KUBEFAAS_NATS_STREAMING_URL="http://defaultKubefaasAuthToken@$(kubectl -n ${KUBEFAAS_NAMESPACE} get svc nats-streaming -o jsonpath='{...ip}:{.spec.ports[0].port}')"

    # ingress controller env
    export INGRESS_CONTROLLER=$(kubectl -n ingress-nginx get svc ingress-nginx -o jsonpath='{...ip}')
}

set_local_build_and_deploy_env() {
    export REPO=kubefaas
    export IMAGE=bundle
    export FETCHER_IMAGE=$REPO/fetcher
    export BUILDER_IMAGE=$REPO/builder
    export PRE_UPGRADE_CHECK_IMAGE=$REPO/pre-upgrade-checks

    export BUILD_ID=$(generate_test_id)
    export TAG=test-${BUILD_ID}
    export PRUNE_INTERVAL=1 # this variable controls the interval to run archivePruner. The unit is in minutes.

    export IMAGE_PULL_POLICY="IfNotPresent"

    export ROUTER_SERVICE_TYPE=NodePort
    export SERVICE_TYPE=NodePort
    export NODE_IP="192.168.1.15"
    export LB_SUPPORT=false

    export KUBEFAAS_NAMESPACE="kubefaas"
    export FUNCTION_NAMESPACE="kubefaas-function"
    export KUBEFAAS_BUILDER_NAMESPACE="kubefaas-builder"

    export CONTROLLER_NODE_PORT=31234
    export ROUTER_NODE_PORT=31235

    TEST_BIN=/tmp/${BUILD_ID}/bin
    mkdir -p ${TEST_BIN}
    export PATH=${TEST_BIN}:${PATH}
}

set_local_test_env() {
    # env
    export KUBEFAAS_HOST=${NODE_IP}:$(kubectl -n ${KUBEFAAS_NAMESPACE} get svc controller -o jsonpath="{.spec.ports[0].nodePort}")
    export KUBEFAAS_URL=http://${NODE_IP}:$(kubectl -n ${KUBEFAAS_NAMESPACE} get svc controller -o jsonpath="{.spec.ports[0].nodePort}")
    export KUBEFAAS_ROUTER=${NODE_IP}:$(kubectl -n ${KUBEFAAS_NAMESPACE} get svc router -o jsonpath="{.spec.ports[0].nodePort}")
    export KUBEFAAS_NATS_STREAMING_URL="http://defaultKubefaasAuthToken@${NODE_IP}:$(kubectl -n ${KUBEFAAS_NAMESPACE} get svc nats-streaming -o jsonpath='{.spec.ports[0].nodePort}')"

    # ingress controller env
#    export INGRESS_CONTROLLER=${NODE_IP}:$(kubectl -n ingress-nginx get svc ingress-nginx -o jsonpath="{.spec.ports[0].nodePort}")
}

setupIngressController() {
    if [ "$LB_SUPPORT" = true ];
    then
        # set up NGINX ingress controller
        kubectl create clusterrolebinding cluster-admin-binding --clusterrole cluster-admin --user $(gcloud config get-value account) || true
        kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.25.1/deploy/static/mandatory.yaml || true
        kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.25.1/deploy/static/provider/cloud-generic.yaml || true
    fi
}

removeIngressController() {
    if [ "$LB_SUPPORT" = true ];
    then
        # set up NGINX ingress controller
        kubectl delete clusterrolebinding cluster-admin-binding || true
        kubectl delete -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.25.1/deploy/static/provider/cloud-generic.yaml || true
        kubectl delete -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/nginx-0.25.1/deploy/static/mandatory.yaml || true
    fi
}

build_and_push_go_mod_cache_image() {
    image_tag=$1
    log_start go_mod_cache_image $image_tag

    if ! docker image ls | grep $image_tag >/dev/null 2>&1 ; then
      docker build -q -t $image_tag -f $ROOT/cmd/bundle/Dockerfile --target godep --build-arg GITCOMMIT=$(getGitCommit) --build-arg BUILDDATE=$(getDate) --build-arg BUILDVERSION=$(getVersion) .
    else
      docker build -q -t $image_tag -f $ROOT/cmd/bundle/Dockerfile --cache-from ${image_tag} --target godep --build-arg GITCOMMIT=$(getGitCommit) --build-arg BUILDDATE=$(getDate) --build-arg BUILDVERSION=$(getVersion) .
    fi

#   docker push $image_tag &
    log_end go_mod_cache_image
}

build_and_push_pre_upgrade_check_image() {
    image_tag=$1
    cache_image=$2
    log_start build_and_push_pre_upgrade_check_image $image_tag

    docker build -q -t $image_tag -f $ROOT/cmd/preupgradechecks/Dockerfile --cache-from ${cache_image} --build-arg GITCOMMIT=$(getGitCommit) --build-arg BUILDDATE=$(getDate) --build-arg BUILDVERSION=$(getVersion) .

#   docker push $image_tag &
    log_end build_and_push_pre_upgrade_check_image
}

build_and_push_kubefaas_bundle() {
    image_tag=$1
    cache_image=$2
    log_start build_and_push_kubefaas_bundle $image_tag

    docker build -q -t $image_tag -f $ROOT/cmd/bundle/Dockerfile --cache-from ${cache_image} --build-arg GITCOMMIT=$(getGitCommit) --build-arg BUILDDATE=$(getDate) --build-arg BUILDVERSION=$(getVersion) .

#   docker push $image_tag &
    log_end build_and_push_kubefaas_bundle
}

build_and_push_fetcher() {
    image_tag=$1
    cache_image=$2
    log_start build_and_push_fetcher $image_tag

    docker build -q -t $image_tag -f $ROOT/cmd/fetcher/Dockerfile --cache-from ${cache_image} --build-arg GITCOMMIT=$(getGitCommit) --build-arg BUILDDATE=$(getDate) --build-arg BUILDVERSION=$(getVersion) .

#   docker push $image_tag &
    log_end build_and_push_fetcher
}


build_and_push_builder() {
    image_tag=$1
    cache_image=$2
    log_start build_and_push_builder $image_tag

    docker build -q -t $image_tag -f $ROOT/cmd/builder/Dockerfile --cache-from ${cache_image} --build-arg GITCOMMIT=$(getGitCommit) --build-arg BUILDDATE=$(getDate) --build-arg BUILDVERSION=$(getVersion) .

#   docker push $image_tag &
    log_end build_and_push_builder
}

build_and_push_env_runtime() {
    env=$1
    image=$2
    image_tag=$3
    variant=$4

    log_start build_and_push_env_runtime.$env $image:$image_tag

    dockerfile="Dockerfile"

    if [ ! -z ${variant} ]; then
        dockerfile=${dockerfile}-${variant}
        image=${image}-${variant}
    fi

    pushd $ROOT/environments/$env/
    docker build -q -t ${image}:${image_tag} . -f ${dockerfile}
    docker tag ${image}:${image_tag} ${image}:latest

#   docker push ${image}:${image_tag} &
#   docker push ${image}:latest &
    popd
    log_end build_and_push_env_runtime.$env
}

build_and_push_env_builder() {
    env=$1
    image=$2
    image_tag=$3
    builder_image=$4
    variant=$5

    log_start build_and_push_env_builder.$env $image:$image_tag

    dockerfile="Dockerfile"

    if [ ! -z ${variant} ]; then
        dockerfile=${dockerfile}-${variant}
        image=${image}-${variant}
    fi

    pushd ${ROOT}/environments/${env}/builder

    docker build -q -t ${image}:${image_tag} --build-arg BUILDER_IMAGE=${builder_image} . -f ${dockerfile}
    docker tag ${image}:${image_tag} ${image}:latest

#   docker push ${image}:${image_tag} &
#   docker push ${image}:latest &
    popd
    log_end build_and_push_env_builder.$env
}

build_kubefaas_cli() {
    log_start build_kubefaas_cli "kubefaas cli"
    pushd $ROOT/cmd/cli
    go build -ldflags "-X github.com/srcmesh/kubefaas/pkg/info.GitCommit=$(getGitCommit) -X github.com/srcmesh/kubefaas/pkg/info.BuildDate=$(getDate) -X github.com/srcmesh/kubefaas/pkg/info.Version=$(getVersion)" -o ${TEST_BIN}/kubefaas .
    popd
    log_end build_kubefaas_cli
}

clean_crd_resources() {
    kubectl --namespace default get crd| grep -v NAME| grep "fission.io"| awk '{print $1}'|xargs -I@ bash -c "kubectl --namespace default delete crd @"  || true
}

generate_test_id() {
    echo $(cat /dev/urandom | tr -dc 'a-z' | fold -w 6 | head -n 1)
}

wait_for_service() {
    ns=$1
    svc=$2

    while true
    do
        ip=$(kubectl -n $ns get svc $svc -o jsonpath='{...ip}')
        if [ ! -z $ip ]; then
           break
        fi
        echo Waiting for service $svc...
        sleep 5
    done
}
export -f wait_for_service

wait_for_services() {
    ns=$1

    if [ "$LB_SUPPORT" = true ];
    then
        wait_for_service $ns controller
        wait_for_service $ns router
        wait_for_service "ingress-nginx" ingress-nginx

        echo Waiting for service is routable...
        sleep 30
    fi
}
export -f wait_for_services

helm_remove_old_releases() {(set -euo pipefail
    # $1: release name
    # $2: namespace

    helm list --all-namespaces | grep "kubefaas" | awk '{printf "helm delete -n %s %s && kubectl delete ns %s && kubectl delete ns %s-func\n", $2, $1, $2, $2}' | xargs -I@ bash -c @ || true
    clean_crd_resources

    kubectl delete ns ${KUBEFAAS_NAMESPACE} || true
    kubectl delete ns ${FUNCTION_NAMESPACE} || true
    kubectl delete ns ${KUBEFAAS_BUILDER_NAMESPACE} || true

    # deleting ns does take a while after command is issued
    while kubectl get ns| grep "kubefaas-builder"
    do
        sleep 5
    done
)}
export -f helm_remove_old_releases

helm_uninstall_kubefaas() {(set +e
    id=$1

    if [ ! -z ${KUBEFAAS_TEST_SKIP_DELETE:+} ]; then
	    echo "Kubefaas uninstallation skipped"
	    return
    fi

    echo "Uninstalling kubefaas"
    helm delete --purge $id
    kubectl delete ns f-$id || true
)}
export -f helm_uninstall_kubefaas

port_forward_services() {
    id=$1
    ns=f-$id
    svc=$2
    port=$3

    kubectl get pods -l svc="$svc" -o name --namespace $ns | \
        sed 's/^.*\///' | \
        xargs -I{} kubectl port-forward {} $port:$port -n $ns &
}

dump_kubefaas_logs() {
    ns=$1
    fns=$2
    component=$3

    echo --- $component logs ---
    kubectl -n $ns get pod -o name | grep $component | xargs kubectl -n $ns logs
    echo --- end $component logs ---
}

describe_pods_ns() {
    echo "--- describe pods $1---"
    kubectl describe pods -n $1
    echo "--- End describe pods $1 ---"
}

describe_all_pods() {
    ns=$1
    fns=$2
    bns=$3

    describe_pods_ns $ns
    describe_pods_ns $fns
    describe_pods_ns $bns
}

dump_all_kubefaas_resources() {
    ns=$1

    echo "--- All objects in the kubefaas namespace $ns ---"
    kubectl -n $ns get pods -o wide
    echo ""
    kubectl -n $ns get svc
    echo "--- End objects in the kubefaas namespace $ns ---"
}

dump_system_info() {
    log_start dump_system_info "System Info"
    go version
    docker version
    kubectl version
    helm version
    log_end dump_system_info
}

install() {
    ns=${KUBEFAAS_NAMESPACE}
    fns=${FUNCTION_NAMESPACE}
    bns=${KUBEFAAS_BUILDER_NAMESPACE}
    repo=${REPO}
    image=${IMAGE}
    imageTag=${TAG}
    fetcherImage=${FETCHER_IMAGE}
    fetcherImageTag=${TAG}
    controllerNodeport=${CONTROLLER_NODE_PORT}
    routerNodeport=${ROUTER_NODE_PORT}
    pruneInterval=${PRUNE_INTERVAL}
    routerServiceType=${ROUTER_SERVICE_TYPE}
    serviceType=${SERVICE_TYPE}
    preUpgradeCheckImage=$${PRE_UPGRADE_CHECK_IMAGE}

    setupIngressController

    helm dependency update $ROOT/charts/kubefaas-all
    helmVars=repository=$repo,image=$image,imageTag=$imageTag,fetcher.image=$fetcherImage,fetcher.imageTag=$fetcherImageTag,functionNamespace=$fns,controllerPort=$controllerNodeport,routerPort=$routerNodeport,pullPolicy=${IMAGE_PULL_POLICY},analytics=false,debugEnv=true,pruneInterval=$pruneInterval,routerServiceType=$routerServiceType,serviceType=$serviceType,preUpgradeChecksImage=$preUpgradeCheckImage,persistence.enabled=false,prometheus.server.persistentVolume.enabled=false,prometheus.alertmanager.enabled=false,prometheus.kubeStateMetrics.enabled=false,prometheus.nodeExporter.enabled=false,prometheus.server.global.evaluation_interval=2s,prometheus.server.global.scrape_interval=2s,prometheus.server.global.scrape_timeout=1s

    echo "Installing kubefaas"
    kubectl create namespace $ns
    helm install \
         --wait	\
         --timeout 300s \
         --name-template "kubefaas" \
         --set $helmVars \
         --namespace $ns \
         $ROOT/charts/kubefaas-all

    helm status -n ${ns} "kubefaas" | grep STATUS | grep -i deployed
    if [ $? -ne 0 ]; then
        describe_all_pods "${ns}" "${fns}" "${bns}"
        helm_remove_old_releases
        removeIngressController
	      exit 1
    fi

    timeout 150 bash -c "wait_for_services ${ns}"
}

run_test_suites() {
    export FAILURES=0   # for later exit code check
    export TIMEOUT=900  # 15 minutes per test

    set +e

    env

    # run tests without newdeploy in parallel.
    export JOBS=5
    $ROOT/test/run_test.sh \
        $ROOT/test/tests/test_pass.sh \
        $ROOT/test/tests/test_canary.sh \
        $ROOT/test/tests/test_fn_update/test_idle_objects_reaper.sh \
        $ROOT/test/tests/mqtrigger/nats/test_mqtrigger.sh \
        $ROOT/test/tests/mqtrigger/nats/test_mqtrigger_error.sh \
        $ROOT/test/tests/mqtrigger/kafka/test_kafka.sh \
        $ROOT/test/tests/test_annotations.sh \
        $ROOT/test/tests/test_archive_pruner.sh \
        $ROOT/test/tests/test_backend_poolmgr.sh \
        $ROOT/test/tests/test_buildermgr.sh \
        $ROOT/test/tests/test_env_vars.sh \
        $ROOT/test/tests/test_environments/test_python_env.sh \
        $ROOT/test/tests/test_function_test/test_fn_test.sh \
        $ROOT/test/tests/test_function_update.sh \
        $ROOT/test/tests/test_obj_create_in_diff_ns.sh \
        $ROOT/test/tests/test_internal_routes.sh \
        $ROOT/test/tests/test_logging/test_function_logs.sh \
        $ROOT/test/tests/test_node_hello_http.sh \
        $ROOT/test/tests/test_package_command.sh \
        $ROOT/test/tests/test_package_checksum.sh \
        $ROOT/test/tests/test_pass.sh \
        $ROOT/test/tests/test_router_cache_invalidation.sh \
        $ROOT/test/tests/test_specs/test_spec.sh \
        $ROOT/test/tests/test_specs/test_spec_multifile.sh \
        $ROOT/test/tests/test_specs/test_spec_merge/test_spec_merge.sh \
        $ROOT/test/tests/test_specs/test_spec_archive/test_spec_archive.sh \
        $ROOT/test/tests/test_environments/test_tensorflow_serving_env.sh \
        $ROOT/test/tests/test_environments/test_go_env.sh \
        $ROOT/test/tests/test_huge_response/test_huge_response.sh \
        $ROOT/test/tests/test_kubectl/test_kubectl.sh \
        $ROOT/test/tests/test_environments/test_java_builder.sh \
        $ROOT/test/tests/test_environments/test_java_env.sh \
        $ROOT/test/tests/test_fn_update/test_configmap_update.sh \
        $ROOT/test/tests/test_fn_update/test_nd_pkg_update.sh \
        $ROOT/test/tests/test_fn_update/test_secret_update.sh \
        $ROOT/test/tests/test_secret_cfgmap/test_secret_cfgmap.sh \
        $ROOT/test/tests/test_fn_update/test_scale_change.sh \
        $ROOT/test/tests/test_backend_newdeploy.sh \
        $ROOT/test/tests/test_fn_update/test_resource_change.sh \
        $ROOT/test/tests/test_fn_update/test_env_update.sh \
        $ROOT/test/tests/test_fn_update/test_poolmgr_nd.sh
        # $ROOT/test/tests/test_ingress.sh \
    FAILURES=$?

    set -e

    # dump test logs
    # TODO: the idx does not match seq number in recap.
    idx=1
    log_files=$(find $ROOT/test/logs/ -name '*.log')
    for log_file in $log_files; do
        test_name=${log_file#$ROOT/test/logs/}
        log_start run_test.$idx $test_name
        echo "========== start $test_name =========="
        cat $log_file
        echo "========== end $test_name =========="
        log_end run_test.$idx
        idx=$((idx+1))
    done
}

cleanup() {
    helm_remove_old_releases
    removeIngressController
}

check_result() {
    ns=${KUBEFAAS_NAMESPACE}
    fns=${FUNCTION_NAMESPACE}
    bns=${KUBEFAAS_BUILDER_NAMESPACE}

#    dump_all_kubefaas_resources $ns
#    dump_kubefaas_logs $ns $fns controller
#    dump_kubefaas_logs $ns $fns router
#    dump_kubefaas_logs $ns $fns buildermgr
#    dump_kubefaas_logs $ns $fns executor
#    dump_kubefaas_logs $ns $fns storagesvc
#    dump_kubefaas_logs $ns $fns mqtrigger-nats-streaming

    if [ $FAILURES -ne 0 ]
    then
        # Commented out due to Travis-CI log length limit
        # describe each pod in kubefaas ns and function namespace
        # describe_all_pods ${build_id}
	      exit 1
    fi
}

wait_for_CI_cluster() {
    while true; do
        # ensure that gke cluster is now free for testing

        previous_build_id=$(kubectl --namespace default get configmap in-test --ignore-not-found -o=jsonpath='{.metadata.labels.buildID}')

        if [[ ! -z ${previous_build_id} ]]; then

            build_state=$(curl -s -X GET http://${DRONE_SYSTEM_HOSTNAME}/api/repos/${CI_REPO_NAME}/builds/${previous_build_id} \
            -H "Authorization: Bearer ${CI_API_TOKEN}" | python3 -c "import sys,json; print(json.load(sys.stdin)['status'])")

            # If previous build state is not equal to "running" or the previous build id
            # equals to the current build ID means the previous build is end or restart.
            # We can remove the configmap safely and start next k8s test safely.
            if [[ ${CI_BUILD_NUMBER} == ${previous_build_id} ]] || [[ $build_state != "running" ]]; then
                kubectl --namespace default delete configmap -l buildID=${previous_build_id}
            fi
        fi

        created=$(kubectl --namespace default create configmap in-test|grep "created"||true)
        if [[ -z $created ]]; then
            echo "Cluster is now in used. Retrying after 15 seconds..."
            sleep 15
            continue
        fi
        kubectl --namespace default label configmap in-test buildID=${CI_BUILD_NUMBER}
        break
    done
}

build_images_and_cli() {
    build_and_push_go_mod_cache_image ${REPO}/go-mod-image-cache
    build_and_push_kubefaas_bundle ${REPO}/$IMAGE:$TAG ${REPO}/go-mod-image-cache
    build_and_push_pre_upgrade_check_image $PRE_UPGRADE_CHECK_IMAGE:$TAG ${REPO}/go-mod-image-cache
    build_and_push_fetcher $FETCHER_IMAGE:$TAG ${REPO}/go-mod-image-cache
    build_and_push_builder $BUILDER_IMAGE:$TAG ${REPO}/go-mod-image-cache

    build_kubefaas_cli
}

build_env_images() {
    export PYTHON_RUNTIME_IMAGE=${REPO}/python-env:${TAG}
    export PYTHON_BUILDER_IMAGE=${REPO}/python-builder:${TAG}
    export JVM_RUNTIME_IMAGE=${REPO}/jvm-env:${TAG}
    export JVM_BUILDER_IMAGE=${REPO}/jvm-builder:${TAG}
    export NODE_RUNTIME_IMAGE=${REPO}/node-env:${TAG}
    export NODE_BUILDER_IMAGE=${REPO}/node-builder:${TAG}
    export TS_RUNTIME_IMAGE=${REPO}/tensorflow-serving-env:${TAG}

    go_variant="1.12"
    export GO_RUNTIME_IMAGE=${REPO}/go-env-${go_variant}:${TAG}
    export GO_BUILDER_IMAGE=${REPO}/go-builder-${go_variant}:${TAG}

    build_and_push_env_runtime python ${REPO}/python-env $TAG ""
    build_and_push_env_runtime jvm ${REPO}/jvm-env $TAG ""
    build_and_push_env_runtime go ${REPO}/go-env $TAG "1.12"
    build_and_push_env_runtime tensorflow-serving ${REPO}/tensorflow-serving-env $TAG ""
    build_and_push_env_runtime nodejs ${REPO}/node-env $TAG ""

    build_and_push_env_builder python ${REPO}/python-builder $TAG $BUILDER_IMAGE:$TAG ""
    build_and_push_env_builder jvm ${REPO}/jvm-builder $TAG $BUILDER_IMAGE:$TAG ""
    build_and_push_env_builder go ${REPO}/go-builder $TAG $BUILDER_IMAGE:$TAG "1.12"
    build_and_push_env_builder nodejs ${REPO}/node-builder $TAG $BUILDER_IMAGE:$TAG ""
}


# if [ $# -lt 2 ]
# then
#     echo "Usage: test.sh [image] [imageTag]"
#     exit 1
# fi
# install_and_test $1 $2
