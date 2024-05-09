// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package harvester

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	harvester "github.com/drewmullen/harvester-go-sdk"
)

// This is a definition of a builder step and should implement multistep.Step
type StepSourceBase struct {
	Name string
}

// Run should execute the purpose of this step
func (s *StepSourceBase) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	client := state.Get("client").(*harvester.APIClient)
	auth := state.Get("auth").(context.Context)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	// if download & checksum, checkImageExists() -> checkImageChecksum()
	// if download & !checksum, checkImageExists()
	// 	 exists: err
	// 	 !exists: downloadImage()
	// if !download, checkImageExists()

	desiredState := int32(100)
	timeout := 2 * time.Minute
	namespace := c.HarvesterNamespace
	url := c.BuilderSource.URL
	ostype := c.BuilderSource.OSType
	sourceName := c.BuilderSource.Name
	var displayName string
	if c.BuilderSource.DisplayName == "" {
		displayName = sourceName
	} else {
		displayName = c.BuilderSource.DisplayName
	}

	annotations := map[string]string{
		"harvesterhci.io/storageClassName": "harvester-longhorn",
	}
	labels := map[string]string{
		"harvesterhci.io/image-type": "raw_qcow2",
		"harvesterhci.io/os-type":    ostype,
	}
	spec := harvester.HarvesterhciIoV1beta1VirtualMachineImageSpec{
		// Description: &desc,
		DisplayName: displayName,
		SourceType:  "download",
		Url:         &url,
	}
	if c.BuilderSource.Checksum != "" {
		spec.Checksum = &c.BuilderSource.Checksum
	}
	img := &harvester.HarvesterhciIoV1beta1VirtualMachineImage{
		ApiVersion: &ApiVersionHarvesterKey,
		Kind:       &KindVirtualMachineImage,
		Metadata: &harvester.K8sIoV1ObjectMeta{
			Name:        &sourceName,
			Annotations: &annotations,
			Labels:      &labels,
			Namespace:   &namespace,
		},
		Spec: spec,
	}

	req := client.ImagesAPI.CreateNamespacedVirtualMachineImage(auth, namespace)
	req = req.HarvesterhciIoV1beta1VirtualMachineImage(*img)
	_, _, err := client.ImagesAPI.CreateNamespacedVirtualMachineImageExecute(req)

	if err != nil {
		ui.Say(fmt.Sprintf("Error creating image: %v", err))
	}

	ui.Say(fmt.Sprintf("Beginning download of image %v...", sourceName))
	err = waitForImageDownload(desiredState, sourceName, c.HarvesterNamespace, *client, auth, timeout, ui)

	if err != nil {
		err := fmt.Errorf("error waiting for image, %v, to finish downloading %s%%: %v", sourceName, string(desiredState), err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Download complete for image %s!", sourceName))

	return multistep.ActionContinue
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepSourceBase) Cleanup(_ multistep.StateBag) {
	// Nothing to clean
}
