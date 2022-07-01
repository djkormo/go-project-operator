#!/bin/bash

export IMG=docker.io/djkormo/go-project-operator:v0.0.6
export NAMESPACE=project-operator
export TAG=v0.0.6

echo "IMG: $IMG"
echo "NAMESPACE: $NAMESPACE"

cd config/manager
kustomize edit set image controller=${IMG}
kustomize edit set namespace "${NAMESPACE}"
cd ../../

cd config/default
kustomize edit set namespace "${NAMESPACE}"
cd ../../
