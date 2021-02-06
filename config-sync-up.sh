#!/bin/bash -xe

gsutil cp gs://config-management-release/released/latest/config-sync-operator.yaml config-sync-operator.yaml
kubectl apply -f config-sync-operator.yaml
kubectl create secret generic git-creds  --namespace=config-management-system  --from-file=ssh=config-sync
cat > config-sync.yaml <<EOF
apiVersion: configmanagement.gke.io/v1
kind: ConfigManagement
metadata:
  name: config-management
spec:
  # clusterName is required and must be unique among all managed clusters
  clusterName: kcc-minikube
  sourceFormat: unstructured
  git:
    syncRepo: https://source.developers.google.com/p/${DEST_PROJECT}/r/kcc-demo
    syncBranch: main
    secretType: gcenode
    policyDir: infra
EOF
kubectl apply -f config-sync.yaml
