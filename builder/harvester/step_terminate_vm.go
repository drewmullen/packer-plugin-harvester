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
type StepTerminateVM struct {
	Name string
}

// Run should execute the purpose of this step
func (s *StepTerminateVM) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	client := state.Get("client").(*harvester.APIClient)
	auth := state.Get("auth").(context.Context)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	if state.Get("Name") == nil {
		ui.Error("VM name is nil")
		return multistep.ActionHalt
	}
	name := state.Get("Name").(string)
	delReq := client.VirtualMachinesAPI.DeleteNamespacedVirtualMachine(auth, name, c.HarvesterNamespace)
	delReq = delReq.K8sIoV1DeleteOptions(harvester.K8sIoV1DeleteOptions{})

	_, _, err := client.VirtualMachinesAPI.DeleteNamespacedVirtualMachineExecute(delReq)

	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting VM: %v", err))
	}

	ui.Say(fmt.Sprintf("The VM, %v, has been terminated", name))

	return multistep.ActionContinue
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepTerminateVM) Cleanup(_ multistep.StateBag) {
	// Nothing to clean
}
