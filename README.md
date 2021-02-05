# gcloud Declarative Demo

1. Create source project and resources

```shell
export SOURCE_PROJECT=vic-gcloud-export-source-7
gcloud projects create --folder=301779790514 ${SOURCE_PROJECT}
gcloud beta billing projects link --billing-account=005196-7B06D5-7D3824 ${SOURCE_PROJECT}
gcloud config set project ${SOURCE_PROJECT}
gcloud services enable cloudasset.googleapis.com cloudresourcemanager.googleapis.com
./create-resources.sh
```

1. Create destination project.

```shell
export DEST_PROJECT=vic-gcloud-declarative-18
gcloud projects create --folder=301779790514 ${DEST_PROJECT}
gcloud beta billing projects link --billing-account=005196-7B06D5-7D3824 ${DEST_PROJECT}
gcloud config set project ${DEST_PROJECT}
gcloud services enable cloudasset.googleapis.com cloudresourcemanager.googleapis.com compute.googleapis.com iam.googleapis.com sourcerepo.googleapis.com
```

1. Install the Kubernetes Config Connector in Minikube

```shell
./kcc-up.sh
```

1. Export the resources to KRM format

```shell
mkdir -p kcc-demo/infra
rm -rf kcc-demo/infra*
# Install the Kubernetes config-connector binary
sudo apt-get install -y google-cloud-sdk-config-connector
gcloud alpha asset bulk-export --path kcc-demo/infra/ --project ${SOURCE_PROJECT}
```

1. Compile the santization function

```shell
cd fn/sanitize-bulk-export/
go build -o ../../sanitize-bulk-export
cd ../..
```

1. Initialize the kpt package

```shell
cd kcc-demo
kpt pkg init .
kpt live init .
# Copy in the kpt declarative funciton
sed s/DEST_PROJECT/${DEST_PROJECT}/g ../fn/sanitize.yaml > sanitize.yaml
sed -i s/SOURCE_PROJECT/${SOURCE_PROJECT}/g sanitize.yaml
cd ..
```

1. Inspect the configuration with --dry-run

```shell
kpt fn run kcc-demo/infra --enable-exec --dry-run
```

1. Ensure KCC is up and running

```shell
kubectl wait -n cnrm-system --for=condition=Ready pod --all
```

1. Apply the config

```shell
kpt fn run kcc-demo/infra --enable-exec
# Workaround project issue b/178745928
rm kcc-demo/infra/project_*.yaml
rm kcc-demo/infra/iam*.yaml
```

## Setting up GitOps

1. Create the git repo in Cloud Source Repositories

```shell
gcloud source repos create kcc-demo
ssh-keygen -t rsa -f config-sync
```

1. Add your SSH public key by visiting the SSH key page:

    https://source.cloud.google.com/user/ssh_keys

1. Push your config to the repo

```shell
cd kcc-demo
git init .
git checkout -b main
git add .
git commit -m "Initial commit"
git remote add origin https://source.developers.google.com/p/${DEST_PROJECT}/r/kcc-demo
git push origin main
cd ..
```

1. Install and configure Config Sync

```shell
gsutil cp gs://config-management-release/released/latest/config-sync-operator.yaml config-sync-operator.yaml
kubectl apply -f config-sync-operator.yaml
kubectl create secret generic git-creds  --namespace=config-management-system  --from-file=ssh=config-syncsecret/git-creds created
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
```

## Cleanup

```shell
minikube delete --profile kcc
gcloud projects delete ${DEST_PROJECT}
gcloud projects delete ${SOURCE_PROJECT}
rm -rf kcc-demo
```
