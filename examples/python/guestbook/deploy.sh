#!/bin/sh

set -e

kubectl create -f redis.yaml

if [ -z "$KUBEFAAS_URL" ]
then
    echo "Need $KUBEFAAS_URL set to a kubefaas controller address"
    exit 1
fi

# Create python env if it doesn't exist
kubefaas env get --name python || kubefaas env create --name python --image kubefaas/python-env

# Register functions and routes with kubefaas
kubefaas function create --name guestbook-get --env python --code get.py --url /guestbook --method GET
kubefaas function create --name guestbook-add --env python --code add.py --url /guestbook --method POST
