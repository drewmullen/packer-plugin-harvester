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
type StepCreateVM struct {
	Name string
}

// Run should execute the purpose of this step
func (s *StepCreateVM) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	client := state.Get("client").(*harvester.APIClient)
	auth := state.Get("auth").(context.Context)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	volName := state.Get("volumeName").(string)

	vm := vmTemplate(c, volName)

	req := client.VirtualMachinesAPI.CreateNamespacedVirtualMachine(auth, c.HarvesterNamespace)

	req = req.KubevirtIoApiCoreV1VirtualMachine(*vm)
	vm, _, err := client.VirtualMachinesAPI.CreateNamespacedVirtualMachineExecute(req)

	// TODO: Gracefully fail if VM with name already exists
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
func (s *StepCreateVM) Cleanup(state multistep.StateBag) {
	client := state.Get("client").(*harvester.APIClient)
	auth := state.Get("auth").(context.Context)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	if state.Get("Name") == nil {
		ui.Error("Cannot cleanup VM, name is nil")
		return
	}
	name := state.Get("Name").(string)
	delReq := client.VirtualMachinesAPI.DeleteNamespacedVirtualMachine(auth, name, c.HarvesterNamespace)
	delReq = delReq.K8sIoV1DeleteOptions(harvester.K8sIoV1DeleteOptions{})

	_, _, err := client.VirtualMachinesAPI.DeleteNamespacedVirtualMachineExecute(delReq)

	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting VM: %v", err))
		return
	}

	waitForVMStateDestroy(name, c.HarvesterNamespace, *client, auth, 10*time.Minute, ui)

	ui.Say(fmt.Sprintf("The VM, %v, has been terminated", name))

}

func vmTemplate(c *Config, volName string) *harvester.KubevirtIoApiCoreV1VirtualMachine {
	return &harvester.KubevirtIoApiCoreV1VirtualMachine{
		ApiVersion: &ApiVersionKubevirt,
		Kind:       &KindVirtualMachine,
		Metadata: &harvester.K8sIoV1ObjectMeta{
			Annotations: &map[string]string{
				"harvesterhci.io/vmRunStrategy":            "RerunOnFailure",
				"kubevirt.io/latest-observed-api-version":  "v1",
				"kubevirt.io/storage-observed-api-version": "v1alpha3",
				"network.harvesterhci.io/ips":              "[]",
			},
			Labels: &map[string]string{
				"harvesterhci.io/creator":      "harvester",
				"harvesterhci.io/os":           "linux",
				"harvesterhci.io/vmName":       "runner",
				"tag.harvesterhci.io/ssh-user": "ubuntu",
			},
			GenerateName: &c.BuilderConfiguration.NamePrefix,
			Namespace:    &c.BuilderConfiguration.Namespace,
		},
		Spec: harvester.KubevirtIoApiCoreV1VirtualMachineSpec{
			RunStrategy: &VirtualMachineSpecRunStrategy,
			Template: harvester.KubevirtIoApiCoreV1VirtualMachineInstanceTemplateSpec{
				Metadata: &harvester.K8sIoV1ObjectMeta{
					Annotations: &map[string]string{
						"harvesterhci.io/waitForLeaseInterfaceNames": "[\"nic-1\"]",
					},
					Labels: &map[string]string{
						"harvesterhci.io/creator":      "packer-plugin-terraform",
						"harvesterhci.io/vmName":       "test",
						"tag.harvesterhci.io/ssh-user": "ubuntu",
					},
				},
				Spec: &harvester.KubevirtIoApiCoreV1VirtualMachineInstanceSpec{
					Affinity: &harvester.K8sIoV1Affinity{
						NodeAffinity: &harvester.K8sIoV1NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &harvester.K8sIoV1NodeSelector{
								NodeSelectorTerms: []harvester.K8sIoV1NodeSelectorTerm{
									{
										MatchExpressions: []harvester.K8sIoV1NodeSelectorRequirement{
											{
												Key:      "network.harvesterhci.io/mgmt",
												Operator: "In",
												Values:   []string{"true"},
											},
										},
									},
								},
							},
						},
						PodAntiAffinity: &harvester.K8sIoV1PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []harvester.K8sIoV1WeightedPodAffinityTerm{
								{
									PodAffinityTerm: harvester.K8sIoV1PodAffinityTerm{
										LabelSelector: &harvester.K8sIoV1LabelSelector{
											MatchExpressions: []harvester.K8sIoV1LabelSelectorRequirement{
												{
													Key:      "harvesterhci.io/creator",
													Operator: "Exists",
												},
											},
										},
										TopologyKey: "kubernetes.io/hostname",
									},
									Weight: 100,
								},
							},
						},
					},
					Domain: harvester.KubevirtIoApiCoreV1DomainSpec{
						Cpu: &harvester.KubevirtIoApiCoreV1CPU{
							Cores: &c.BuilderConfiguration.CPU,
						},
						Devices: harvester.KubevirtIoApiCoreV1Devices{
							Disks: []harvester.KubevirtIoApiCoreV1Disk{
								{
									BootOrder: toInt32Ptr(1),
									Disk: &harvester.KubevirtIoApiCoreV1DiskTarget{
										Bus: toStringPtr("virtio"),
									},
									Name: "rootdisk",
								},
								{
									Disk: &harvester.KubevirtIoApiCoreV1DiskTarget{
										Bus: toStringPtr("virtio"),
									},
									Name: "cloudinitdisk",
								},
							},
							Interfaces: []harvester.KubevirtIoApiCoreV1Interface{
								{
									Bridge:     map[string]interface{}{},
									MacAddress: toStringPtr("1e:31:ed:d3:83:9c"),
									Model:      toStringPtr("virtio"),
									Name:       "nic-1",
								},
							},
						},
						Features: &harvester.KubevirtIoApiCoreV1Features{
							Acpi: &harvester.KubevirtIoApiCoreV1FeatureState{},
							Smm: &harvester.KubevirtIoApiCoreV1FeatureState{
								Enabled: toBoolPtr(true),
							},
						},
						Firmware: &harvester.KubevirtIoApiCoreV1Firmware{
							Bootloader: &harvester.KubevirtIoApiCoreV1Bootloader{
								Efi: &harvester.KubevirtIoApiCoreV1EFI{
									SecureBoot: toBoolPtr(true),
								},
							},
						},
						Machine: &harvester.KubevirtIoApiCoreV1Machine{
							Type: toStringPtr("q35"),
						},
						Resources: &harvester.KubevirtIoApiCoreV1ResourceRequirements{
							Requests: map[string]string{
								"memory": c.BuilderConfiguration.Memory,
							},
						},
					},
					EvictionStrategy: toStringPtr("LiveMigrate"),
					Hostname:         toStringPtr("test"),
					Networks: []harvester.KubevirtIoApiCoreV1Network{
						{
							Multus: &harvester.KubevirtIoApiCoreV1MultusNetwork{
								NetworkName: "harvester-public/lab",
							},
							Name: "nic-1",
						},
					},
					TerminationGracePeriodSeconds: toInt64Ptr(120),
					Volumes: []harvester.KubevirtIoApiCoreV1Volume{
						{
							Name: "rootdisk",
							PersistentVolumeClaim: &harvester.KubevirtIoApiCoreV1PersistentVolumeClaimVolumeSource{
								ClaimName: volName,
							},
						},
						{
							CloudInitNoCloud: &harvester.KubevirtIoApiCoreV1CloudInitNoCloudSource{
								NetworkDataSecretRef: &harvester.K8sIoV1LocalObjectReference{
									Name: toStringPtr("packer"),
								},
								SecretRef: &harvester.K8sIoV1LocalObjectReference{
									Name: toStringPtr("packer"),
								},
							},
							Name: "cloudinitdisk",
						},
					},
				},
			},
		},
	}
}

func toStringPtr(s string) *string {
	return &s
}

func toBoolPtr(b bool) *bool {
	return &b
}

func toInt64Ptr(i int64) *int64 {
	return &i
}

func toInt32Ptr(i int32) *int32 {
	return &i
}
