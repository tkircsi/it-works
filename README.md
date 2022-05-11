# It-Works (operator)
It-Works is a simple k8s operator which makes a 
deployment app, a service and an ingress rule. The application
is accessible at the `host` address via HTTP and HTTPS.
## It-Works CRD
The CRD specifies 3 paramters
1. replicas: number of replicas
2. host: host where the application is accessible
3. image: container image (and tag) (eg. nginx:latest )
```
apiVersion: webapp.tkircsi/v1alpha1
kind: ItWorks
metadata:
  name: it-works
  namespace: default
spec:
  replicas: 2
  host: it.works.io
  image: nginx:latest

```
## Installation (minikube)
### Start minikube
`minikube start`
### Enable ingress addon
`minikube addons enable ingress`

Check the ingress addon is enabled

`minikube addons list`

## Check kube context
Check the current context is minikube

`kubectl config get-contexts`
## Install cert-manager
We need cert-manager to issue certificates. Now we will use
self-signed certificates.

`kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml`

Check cert-manager deployment

`kubectl get all -n cert-manager`

## Install cert-manager issuers
```
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: selfsigned-issuer
  namespace: default
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-cluster-issuer
spec:
  selfSigned: {}
```

Check cert-manager issuers.

`kubectl get issuers.cert-manager.io`

`kubectl get clusterissuers.cert-manager.io`

## Load images into minikube local repo
`minikube image load gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0`

`minikube image load tkircsi/it-works:v0.0.2`

## Deploy the It-Works operator
`kubectl apply -f https://raw.githubusercontent.com/tkircsi/it-works/main/it-works-operator.yaml`

Check the operator

`kubectl get all -n it-works-systemkubectl get all -n it-works-system`

## Deploy the the app (CRD)
`kubectl apply -f https://raw.githubusercontent.com/tkircsi/it-works/main/itworks.yml`

Check the deployment and service

`kubectl get all`

Check the tls-secret

`kubectl get secrets`

Check the certificates

`kubectl get certificate`

## Add host name to the hosts file
Add the following line to the host file

`127.0.0.1 it.works.io`

## Testing the application
Start minikube tunnel

`minikube tunnel`

Open a browser i.e. Chrome and go to 
the `https://it.works.io` address. If the browser complains
 about the self-signed certificate in Chrome type 'thisisunsafe'
anywhere in the Chrome window.

We should see the nginx default page if everything is fine.




