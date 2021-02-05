#!/bin/bash -xe

export OLD_PROJECT=$(gcloud config get-value project)
gcloud config set project $SOURCE_PROJECT

# Enable GCE API
gcloud services enable compute.googleapis.com
gcloud services enable cloudasset.googleapis.com

# Cleanup project default project resources
gcloud compute firewall-rules delete default-allow-icmp default-allow-internal default-allow-rdp default-allow-ssh --quiet || true
gcloud compute networks delete default --quiet || true

# Create an instance
gcloud compute networks create shiny --subnet-mode=custom
gcloud compute networks subnets create shiny-us-central1 --network shiny --range 10.100.0.0/20 --region us-central1
gcloud compute firewall-rules create allow-remote --network shiny --allow tcp:22,tcp:3389,icmp
gcloud compute instances create shiny --subnet shiny-us-central1 --zone=us-central1-a --no-address --machine-type f1-micro

# Some buckets
gsutil mb gs://${SOURCE_PROJECT}-gcloud-declarative-1
gsutil mb gs://${SOURCE_PROJECT}-gcloud-declarative-2

gcloud config set project ${OLD_PROJECT}