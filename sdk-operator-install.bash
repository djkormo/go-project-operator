#!/bin/bash

# Kustomize 

curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash

chmod a+x kustomize
sudo mv kustomize /usr/local/bin/kustomize



# Kubebuilder 

#curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)
#sudo chmod +x kubebuilder && mv kubebuilder /usr/local/bin/kubebuilder

# Operator SDK

git clone https://github.com/operator-framework/operator-sdk
cd operator-sdk
git checkout v1.19.x
make install