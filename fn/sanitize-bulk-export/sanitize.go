// main.go
package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var oldProject string
var newProject string

func main() {
	resourceList := &framework.ResourceList{}

	cmd := framework.Command(resourceList, func() error {
		// cmd.Execute() will parse the ResourceList.functionConfig into
		// cmd.Flags from the ResourceList.functionConfig.data field.
		for i := range resourceList.Items {
			// modify the resources using the kyaml/yaml library:
			// https://pkg.go.dev/sigs.k8s.io/kustomize/kyaml/yaml

			// Set the project-id annotation
			resourceList.Items[i].SetAnnotations(map[string]string{"cnrm.cloud.google.com/project-id": newProject})
			if oldProject == "" {
				panic("Must set oldProject value")
			}
			if newProject == "" {
				panic("Must set newProject value")
			}
			// Replace all references to the old project
			searchAndReplace(resourceList.Items[i], oldProject, newProject)

			meta, err := resourceList.Items[i].GetMeta()
			if err != nil {
				return err
			}

			switch meta.Kind {
			case "ComputeInstance":
				err := sanitizeComputeInstance(resourceList.Items[i])
				if err != nil {
					return err
				}
			case "IAMServiceAccount":
				// TODO Default service account gets rejected because name starts with integers
				///     This causes the iampolicy to fail to apply as well
				beginsWithNumber := regexp.MustCompile(`^[0-9].*`)
				if beginsWithNumber.MatchString(meta.Name) {
					meta.Name = "sa-" + meta.Name
				}
			case "IAMPolicy":
				// TODO should rename newProject->destProject
				if meta.Name == fmt.Sprintf("%s-iampolicy", newProject) {
					// TODO SWAP OUT ASAP
					ownerList := fmt.Sprintf("members:\n- serviceAccount:kcc-minikube@%s.iam.gserviceaccount.com\nrole: roles/owner", newProject)
					owners, err := yaml.Parse(ownerList)
					if err != nil {
						// TODO wrap these errors with Errorf
						return err
					}
					err = resourceList.Items[i].PipeE(yaml.LookupCreate(yaml.MappingNode, "spec", "bindings", "[=-roles/owner]"), yaml.FieldSetter{Value: owners})
					return err

				}
			}

		}
		return nil
	})
	cmd.Flags().StringVar(&oldProject, "oldProject", "", "flag value")
	cmd.Flags().StringVar(&newProject, "newProject", "", "flag value")
	serviceAccountDefault := fmt.Sprintf("kcc-minikube@%s.iam.gserviceaccount.com", newProject)
	cmd.Flags().StringVar(&newProject, "serviceAccount", serviceAccountDefault, "flag value")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func searchAndReplace(resource *yaml.RNode, old string, new string) error {
	// Fix up project name and project ID
	jsonBytes, err := resource.MarshalJSON()
	jsonBytes = []byte(strings.ReplaceAll(string(jsonBytes), old, new))
	resource.UnmarshalJSON(jsonBytes)
	return err
}

func sanitizeComputeInstance(resource *yaml.RNode) error {
	// Remove private IP address
	err := resource.PipeE(yaml.Lookup("spec", "networkInterface", "[name=nic0]"), yaml.Clear("networkIp"))
	// Clear service account
	err = resource.PipeE(yaml.Lookup("spec", "serviceAccount"), yaml.Clear("serviceAccountRef"))
	// Clear boot disk source reference
	err = resource.PipeE(yaml.Lookup("spec", "bootDisk"), yaml.Clear("sourceDiskRef"))
	return err
}
