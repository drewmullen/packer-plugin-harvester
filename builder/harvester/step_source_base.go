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

	desiredState := int32(100)
	timeout := 2 * time.Minute
	namespace := c.HarvesterNamespace
	url := c.BuilderSource.URL
	checkSum := c.BuilderSource.Checksum
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
	
	preExistingImg, err := checkImageExists(client, auth, displayName, namespace)
	if err != nil {
		//TODO: if error is image does not exist then output info and continue, else error and halt
		if (preExistingImg == harvester.HarvesterhciIoV1beta1VirtualMachineImage{}) {
			if url != "" {
				ui.Say("INFO: image does not exist. continuing... ")
			} else {
				ui.Error("ERROR: image does not exist and no download url provided")
				return multistep.ActionHalt
			}

		}
	} else {
		if url != "" && checkSum != "" {
			if *preExistingImg.Spec.Checksum == "" {
				ui.Error(fmt.Sprintf("ERROR: checksum not set for pre-existing image %s. Unable to compare images.", displayName))
				return multistep.ActionHalt
			}

			if checkSum != *preExistingImg.Spec.Checksum {
				ui.Error("ERROR: Image checksums do not match. either erase prior image or rename new image.")
				return multistep.ActionHalt
			}
			ui.Say("INFO: image already exists and checksums match. skipping download")
			return multistep.ActionContinue

		} else if url != "" && checkSum == "" {
			ui.Error(fmt.Sprintf("ERROR: image with matching name, %s, already exists and no checksum provided. unable to compare checksums. either provide a checksum or change the name of the image to be unique", displayName))
			return multistep.ActionHalt
		} else if url == "" {
			ui.Say("INFO: image already exists skipping download")
			return multistep.ActionContinue
		}

	}

	req := client.ImagesAPI.CreateNamespacedVirtualMachineImage(auth, namespace)
	req = req.HarvesterhciIoV1beta1VirtualMachineImage(*img)
	_, resp, err := client.ImagesAPI.CreateNamespacedVirtualMachineImageExecute(req)

	if err != nil {
		ui.Error(fmt.Sprintf("Error creating image: %v \n %v", err, resp))
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Beginning download of image %v...", sourceName))
	err = waitForImageDownload(desiredState, sourceName, c.HarvesterNamespace, *client, auth, timeout, ui)

	if err != nil {
		err := fmt.Errorf("error waiting for image, %v, to finish downloading %s%%: %v", sourceName, string(desiredState), err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Download complete for image %s!", sourceName))

	return multistep.ActionContinue
}

//

func checkImageExists(client *harvester.APIClient, auth context.Context, displayName string, namespace string) (harvester.HarvesterhciIoV1beta1VirtualMachineImage, error) {
	req := client.ImagesAPI.ReadNamespacedVirtualMachineImage(auth, displayName, namespace)
	preExistingImg, _, err := client.ImagesAPI.ReadNamespacedVirtualMachineImageExecute(req)

	if err != nil {
		return harvester.HarvesterhciIoV1beta1VirtualMachineImage{}, err
	}
	if preExistingImg == nil {
		return harvester.HarvesterhciIoV1beta1VirtualMachineImage{}, err
	}
	return *preExistingImg, err
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepSourceBase) Cleanup(_ multistep.StateBag) {
	// Nothing to clean
}
