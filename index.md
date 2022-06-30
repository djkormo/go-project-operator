## Using helm chart

```console
helm repo add djkormo-go-project-operator https://djkormo.github.io/go-project-operator/

helm repo update

helm search repo go-project-operator  --versions

helm install go-project-operator djkormo-go-project-operator/go-project-operator \
  --namespace project-operator --values charts/go-project-operator/values.yaml \
  --create-namespace --dry-run

helm upgrade project-operator djkormo-go-project-operator/go-project-operator \
  --namespace project-operator --values charts/go-project-operator/values.yaml


helm uninstall go-project-operator  --namespace project-operator 

```
