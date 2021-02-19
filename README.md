#  Moving from ClickOps to GitOps for Infrastructure Management

[![Open in Cloud Shell](https://gstatic.com/cloudssh/images/open-btn.png)](https://ssh.cloud.google.com/cloudshell/open?cloudshell_git_repo=https://github.com/viglesiasce/clickops-to-gitops-demo&cloudshell_tutorial=README.md)

1. Set some variables that will be reused throughout the tutorial:

```sh
export SOURCE_PROJECT=vic-gcloud-export-source-8
export DEST_PROJECT=vic-gcloud-destination-8
export PROJECT_CREATION_ARGS='--folder=301779790514'
```

1. Create source project and resources

```sh
gcloud projects create ${PROJECT_CREATION_ARGS} ${SOURCE_PROJECT}
gcloud beta billing projects link --billing-account=005196-7B06D5-7D3824 ${SOURCE_PROJECT}
gcloud config set project ${SOURCE_PROJECT}
gcloud services enable cloudasset.googleapis.com cloudresourcemanager.googleapis.com
./create-resources.sh
```

1. Create destination project.

```sh
gcloud projects create ${PROJECT_CREATION_ARGS} ${DEST_PROJECT}
gcloud beta billing projects link --billing-account=005196-7B06D5-7D3824 ${DEST_PROJECT}
gcloud config set project ${DEST_PROJECT}
gcloud services enable cloudasset.googleapis.com cloudresourcemanager.googleapis.com compute.googleapis.com iam.googleapis.com sourcerepo.googleapis.com
```

1. Install the Kubernetes Config Connector in Minikube

```sh
./kcc-up.sh
```

1. Export the resources to KRM format

```sh
mkdir -p kcc-demo/infra
rm -rf kcc-demo/infra/*
# Install the Kubernetes config-connector binary
echo y | sudo apt-get install -y google-cloud-sdk-config-connector
gcloud alpha resource-config bulk-export --path kcc-demo/infra/ --project ${SOURCE_PROJECT}
```

1. Compile the santization function

```sh
cd fn/sanitize-bulk-export/
go build -o ../../sanitize-bulk-export
cd ../..
```

1. Initialize the kpt package

```sh
cd kcc-demo
kpt pkg init .
kpt live init .
# Copy in the kpt declarative funciton
sed s/DEST_PROJECT/${DEST_PROJECT}/g ../fn/sanitize.yaml > infra/sanitize.yaml
sed -i s/SOURCE_PROJECT/${SOURCE_PROJECT}/g infra/sanitize.yaml
cd ..
```

1. Inspect the configuration with --dry-run

```sh
kpt fn run kcc-demo/infra --enable-exec --dry-run
```

1. Ensure KCC is up and running

```sh
kubectl wait -n cnrm-system --for=condition=Ready pod --all
```

1. Apply the config

```sh
kpt fn run kcc-demo/infra --enable-exec
# Workaround project issue b/178745928
rm kcc-demo/infra/project_*.yaml
rm kcc-demo/infra/iam*.yaml
```

## Setting up GitOps

1. Create the git repo in Cloud Source Repositories

```sh
gcloud source repos create kcc-demo
ssh-keygen -t rsa -f config-sync
```

1. Add your SSH public key by visiting the SSH key page:

```sh
cat config-sync.pub
```

Visit the [Register SSH Key](https://source.cloud.google.com/user/ssh_keys?register=true).

1. Push your config to the repo

```sh
cd kcc-demo
git init .
git checkout -b main
git add .
git commit -m "Initial commit"
git remote add origin https://source.developers.google.com/p/${DEST_PROJECT}/r/kcc-demo
git push --set-upstream origin main
cd ..
```

1. Install and configure Config Sync

```sh
./config-sync-up.sh
```

## Cleanup

```sh
minikube delete --profile kcc
gcloud projects delete ${DEST_PROJECT}
gcloud projects delete ${SOURCE_PROJECT}
rm -rf kcc-demo
```
