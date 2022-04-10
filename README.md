
## 1. Creating template for operator

```
operator-sdk init --domain=djkormo.github.io --repo=github.com/djkormo/go-project-operator --skip-go-version-check

operator-sdk create api --group=project --version=v1alpha1 --kind=Project
```

## 2. Adding fields to our Project CR





Literature:

https://dev.to/ishankhare07/writing-a-simple-kubernetes-controller-in-go-with-kubebuilder-ib8

https://www.techtarget.com/searchitoperations/tutorial/How-to-build-a-Kubernetes-operator


https://book.kubebuilder.io/reference/markers.html

