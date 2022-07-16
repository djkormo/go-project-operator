#!/bin/bash

export IMG=docker.io/djkormo/go-project-operator:0.0.13
export NAMESPACE=project-operator
export TAG=0.0.13

echo "IMG: $IMG"
echo "NAMESPACE: $NAMESPACE"

cd config/manager
kustomize edit set image controller=${IMG}
kustomize edit set namespace "${NAMESPACE}"
cd ../../

cd config/default
kustomize edit set namespace "${NAMESPACE}"
cd ../../
