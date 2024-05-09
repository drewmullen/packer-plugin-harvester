// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package harvester

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	harvester "github.com/drewmullen/harvester-go-sdk"
)

// This is a definition of a builder step and should implement multistep.Step
type StepCreateVolume struct {
	Name string
}

// Run should execute the purpose of this step
func (s *StepCreateVolume) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	client := state.Get("client").(*harvester.APIClient)
	auth := state.Get("auth").(context.Context)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	req := client.VolumesAPI.CreateNamespacedPersistentVolumeClaim(auth, c.HarvesterNamespace)

	claimInput := &harvester.K8sIoV1PersistentVolumeClaim{
		Metadata: &harvester.K8sIoV1ObjectMeta{
			GenerateName: &c.BuilderConfiguration.NamePrefix,
			Annotations: &map[string]string{
				"harvesterhci.io/imageId": fmt.Sprintf("%s/%s", c.HarvesterNamespace, c.BuilderSource.Name),
			},
		},
		Spec: &harvester.K8sIoV1PersistentVolumeClaimSpec{
			AccessModes: []string{"ReadWriteMany"},
			Resources: &harvester.K8sIoV1ResourceRequirements{
				Requests: map[string]string{
					"storage": c.BuilderTarget.VolumeSize,
				},
			},
			StorageClassName: toStringPtr(GetImageStorageClassName(c.BuilderSource.Name)),
			VolumeMode:       toStringPtr("Block"),
		},
	}
	req = req.K8sIoV1PersistentVolumeClaim(*claimInput)
	claim, _, err := req.Execute()

	if err != nil {
		ui.Error(fmt.Sprintf("Error creating volume: %v", err))
	}

	if claim.Metadata.Name == nil || *claim.Metadata.Name == "" {
		ui.Error("Volume name is empty")
		return multistep.ActionHalt
	}
	state.Put("volumeName", *claim.Metadata.Name)

	ui.Say(fmt.Sprintf("Volume %s created and ready for use", *claim.Metadata.Name))

	// Determines that should continue to the next step
	return multistep.ActionContinue
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepCreateVolume) Cleanup(state multistep.StateBag) {
	// Nothing to clean

	client := state.Get("client").(*harvester.APIClient)
	auth := state.Get("auth").(context.Context)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	volumeName := state.Get("volumeName").(string)

	ui.Say(fmt.Sprintf("Deleting volume %s in namespace %s", volumeName, c.HarvesterNamespace))

	req := client.VolumesAPI.DeleteNamespacedPersistentVolumeClaim(auth, volumeName, c.HarvesterNamespace)
	req = req.K8sIoV1DeleteOptions(harvester.K8sIoV1DeleteOptions{})
	_, _, err := req.Execute()

	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting volume: %v", err))
	}
}

func GetImageStorageClassName(imageName string) string {
	return fmt.Sprintf("longhorn-%s", imageName)
}
