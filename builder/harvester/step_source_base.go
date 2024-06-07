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
	"os"
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
	//test := client.VirtualMachinesAPI.ReadNamespacedVirtualMachineInstance(auth, "test", c.HarvesterNamespace)
	//_, resp, err := test.Execute()
	//how would i do this
	
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
	
	tempImg, errCheckSum:= checkImageExists(client, auth, displayName, namespace)
	//check for uninitialized image
	if(tempImg==harvester.HarvesterhciIoV1beta1VirtualMachineImage{}){
		ui.Say(fmt.Sprintf("image is not initialized %v", errCheckSum))
		os.Exit(1)
	}

	if url != "" && checkSum != "" {
		if errCheckSum != nil || *tempImg.Spec.Checksum=="" {
			ui.Say(fmt.Sprintf("image with name %v does not exist", sourceName))
			ui.Say("exiting program...")
			os.Exit(1)
		}

		if checkSum != *tempImg.Spec.Checksum {
			ui.Say(fmt.Sprintf("image with %v already exists and with a different check sum", sourceName))
			os.Exit(1)
		}

	} else if url != "" && checkSum == "" {
		if errCheckSum != nil {
			ui.Say(fmt.Sprintf("image already exists %v", errCheckSum))
			ui.Say("exiting program...")
			os.Exit(1)
		}

	}else if url =="" && checkSum!=""{
		//should have this error out
		ui.Say("has no url but has checksum. provide url to download image")
		if errCheckSum != nil {
			ui.Say(fmt.Sprintf("image does not exist %v", errCheckSum))
			ui.Say("exiting program...")
			os.Exit(1)
		}
		os.Exit(1)

	}else if url == "" && checkSum==""{
		//should have this error out
		ui.Say("No url and no check sum. provide url and checksum to download image")
		if errCheckSum != nil {
			ui.Say(fmt.Sprintf("image does not exist %v", errCheckSum))
			ui.Say("exiting program...")
			os.Exit(1)
		}
		os.Exit(1)
	}
	
	// this is where 409 error is coming from vm still gets created just wont create a new image since we used a pulled down image
	// question is do we want to create when we pull down? this can still fire and everything else works 
	// just wont create the image if we pull one down instead of full creation from scratch
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

//

func checkImageExists(client *harvester.APIClient, auth context.Context, displayName string, namespace string) (harvester.HarvesterhciIoV1beta1VirtualMachineImage, error) {
	test := client.ImagesAPI.ReadNamespacedVirtualMachineImage(auth, displayName, namespace)
	tempimg, _, err := client.ImagesAPI.ReadNamespacedVirtualMachineImageExecute(test)
	
	if err != nil {
		return harvester.HarvesterhciIoV1beta1VirtualMachineImage{}, err
	}
	if tempimg==nil{
		return harvester.HarvesterhciIoV1beta1VirtualMachineImage{}, err
	}
	return *tempimg, err
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepSourceBase) Cleanup(_ multistep.StateBag) {
	// Nothing to clean
}
