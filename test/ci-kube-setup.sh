#!/bin/bash

#
# Download kubectl, save kubeconfig, and ensure we can access the test cluster
#

set -e 

TOOL_DIR=${HOME}/bin

if [ ! -d $TOOL_DIR ]
then
    mkdir -p $TOOL_DIR
fi

# Get staticcheck
STATICCHECK_VERSION=2020.1.3
if [ ! -f $TOOL_DIR/staticcheck ] || (staticcheck -version | grep -v $STATICCHECK_VERSION)
then
    curl -LO https://github.com/dominikh/go-tools/releases/download/${STATICCHECK_VERSION}/staticcheck_linux_amd64.tar.gz
    tar xzvf staticcheck_linux_amd64.tar.gz
    mv staticcheck/staticcheck $TOOL_DIR/staticcheck
fi

# Get helm
HELM_VERSION=3.1.1
if [ ! -f $TOOL_DIR/helm ] || (helm version --client | grep -v $HELM_VERSION)
then
    curl -LO https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz
    tar xzvf helm-*.tar.gz
    mv linux-amd64/helm $TOOL_DIR/helm
fi

# Get kubectl
if [ ! -f $TOOL_DIR/kubectl ]
then
   curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
   chmod +x ./kubectl
   mv kubectl $TOOL_DIR/kubectl
fi

mkdir ${HOME}/.kube

echo $KUBECONFIG_CONTENTS | base64 -d > ${HOME}/.kube/config

if [ ! -f ${HOME}/.kube/config ]
then
    echo "Missing kubeconfig"
    exit 1
fi

kubectl get node
