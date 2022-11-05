#!/bin/bash

sudo apt-get update

sudo apt-get install snap

exit 0


# uninstall existing golang 

sudo rm -rvf /usr/local/go/

# install go 1.17 


VERSION="1.17.8" # go version
ARCH="amd64" # go archicture
curl -O -L "https://golang.org/dl/go${VERSION}.linux-${ARCH}.tar.gz"
ls -l

#Extract the tarball using the tar command:

sudo tar -xf "go${VERSION}.linux-${ARCH}.tar.gz"
ls -l
cd go/
ls -l
cd ..


#Set up the permissions using the chown command/chmod command:
sudo chown -R root:root ./go

sudo rm -f -R /usr/local/go

sudo mv -v go /usr/local

rm -f go*.tar.gz

cd ~


# Kustomize 

curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash

chmod a+x kustomize
sudo mv kustomize /usr/local/bin/kustomize



# Kubebuilder 

cd ~

curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)
sudo chmod +x kubebuilder && sudo mv kubebuilder /usr/local/bin/kubebuilder

# Operator SDK

git clone https://github.com/operator-framework/operator-sdk
cd operator-sdk
git checkout v1.19.x
make install


cd ~

alias k='kubectl'

rm -fr operator-sdk

echo "Component Versions"
kustomize version
kubebuilder version
operator-sdk version
helm version 

minikube start -p aged --kubernetes-version=v1.22.10

