// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package harvester

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	harvester "github.com/drewmullen/harvester-go-sdk"
)

// This is a definition of a builder step and should implement multistep.Step
type StepExportVMImage struct {
	Name string
}

// Run should execute the purpose of this step
func (s *StepExportVMImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	client := state.Get("client").(*harvester.APIClient)
	auth := state.Get("auth").(context.Context)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	// temp for hardcoded payload
	// vmBody := []byte(vmStr)
	// vmObj := &harvester.KubevirtIoApiCoreV1VirtualMachine{}
	// json.Unmarshal(vmBody, vmObj)

	req := client.VirtualMachinesAPI.CreateNamespacedVirtualMachine(auth, c.HarvesterNamespace)

	// req = req.KubevirtIoApiCoreV1VirtualMachine(*vmObj)
	vm, _, err := client.VirtualMachinesAPI.CreateNamespacedVirtualMachineExecute(req)

	if err != nil {
		ui.Error(fmt.Sprintf("Error creating VM: %v", err))
	}

	// could use generateName
	if vm.Metadata.Name == nil {
		ui.Error("VM name is nil")
		return multistep.ActionHalt
	}
	name := *vm.Metadata.Name
	state.Put("Name", *vm.Metadata.Name)

	ui.Say(fmt.Sprintf("Creating builder VM. Name is %v", name))
	ui.Say(fmt.Sprintf("Waiting for VM, %v, to report as \"Running\"", name))

	timeout := 2 * time.Minute
	desiredState := "Running"
	time.Sleep(3 * time.Second)
	err = waitForVMState(desiredState, name, c.HarvesterNamespace, *client, auth, timeout, ui)

	if err != nil {
		err := fmt.Errorf("error waiting for vm, %v, to become %v: %s", name, desiredState, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("VM created and ready for use")

	// Determines that should continue to the next step
	return multistep.ActionContinue
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepExportVMImage) Cleanup(_ multistep.StateBag) {
	// Nothing to clean
}

func waitForVMImageExport(desiredState string, name string, namespace string, client harvester.APIClient, auth context.Context, timeout time.Duration, ui packersdk.Ui) error {
	startTime := time.Now()

	for {
		readReq := client.VirtualMachinesAPI.ReadNamespacedVirtualMachineInstance(auth, name, namespace)
		currentState, _, err := readReq.Execute()
		if err != nil {
			return err
		}

		if *currentState.Status.Phase == desiredState {
			return nil
		}

		if time.Since(startTime) >= timeout {
			return errors.New("timeout waiting for desired state")
		}

		ui.Say("Waiting for VM to be ready...")
		time.Sleep(5 * time.Second) // Adjust the polling interval as needed
	}
}
