#!/bin/bash -xe

# Get a service account for KCC to use in the destination project
gcloud iam service-accounts create kcc-minikube
gcloud projects add-iam-policy-binding ${DEST_PROJECT} --member="serviceAccount:kcc-minikube@${DEST_PROJECT}.iam.gserviceaccount.com"     --role="roles/owner"
gcloud iam service-accounts keys create --iam-account kcc-minikube@${DEST_PROJECT}.iam.gserviceaccount.com key.json

# Start up minikube
export GOOGLE_APPLICATION_CREDENTIALS="key.json"
minikube start --profile kcc --addons=gcp-auth

# Install and configure KCC
kubectl create namespace cnrm-system
kubectl create secret generic gcloud-declarative-kcc --from-file key.json --namespace cnrm-system
gsutil cp gs://configconnector-operator/latest/release-bundle.tar.gz release-bundle.tar.gz
tar zxvf release-bundle.tar.gz
kubectl apply -f operator-system/configconnector-operator.yaml
cat > configconnector.yaml <<EOF
apiVersion: core.cnrm.cloud.google.com/v1beta1
kind: ConfigConnector
metadata:
  # the name is restricted to ensure that there is only ConfigConnector

  # instance installed in your cluster

  name: configconnector.core.cnrm.cloud.google.com
spec:
 mode: cluster
 credentialSecretName: gcloud-declarative-kcc
EOF
kubectl apply -f configconnector.yaml
