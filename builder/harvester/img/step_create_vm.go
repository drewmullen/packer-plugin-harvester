// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package img

import (
	"context"
	"encoding/json"
	"fmt"

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

	// temp for hardcoded payload
	vmBody := []byte(vmStr)
	vmObj := &harvester.KubevirtIoApiCoreV1VirtualMachine{}
	json.Unmarshal(vmBody, vmObj)

	req := client.VirtualMachinesAPI.CreateNamespacedVirtualMachine(auth, c.HarvesterNamespace)

	req = req.KubevirtIoApiCoreV1VirtualMachine(*vmObj)
	vm, _, err := client.VirtualMachinesAPI.CreateNamespacedVirtualMachineExecute(req)

	if err != nil {
		ui.Error(fmt.Sprintf("Error creating VM: %v", err))
	}

	// could use generateName
	if vm.Metadata.Name == nil {
		ui.Error("VM name is nil")
	}
	s.Name = *vm.Metadata.Name

	ui.Say(fmt.Sprintf("The vm name is %v", s.Name))
	// Determines that should continue to the next step
	return multistep.ActionContinue
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepCreateVM) Cleanup(_ multistep.StateBag) {
	// Nothing to clean
}

var vmStr string = `{
	"apiVersion": "kubevirt.io/v1",
	"kind": "VirtualMachine",
	"metadata": {
		"annotations": {
			"harvesterhci.io/vmRunStrategy": "RerunOnFailure",
			"kubevirt.io/latest-observed-api-version": "v1",
			"kubevirt.io/storage-observed-api-version": "v1alpha3",
			"network.harvesterhci.io/ips": "[]",
			"harvesterhci.io/volumeClaimTemplates": "[{\"metadata\":{\"name\":\"packer-build\",\"creationTimestamp\":null,\"annotations\":{\"harvesterhci.io/imageId\":\"harvester-public/ubuntu-22\",\"terraform-provider-harvester-auto-delete\":\"true\"}},\"spec\":{\"accessModes\":[\"ReadWriteMany\"],\"resources\":{\"requests\":{\"storage\":\"100Gi\"}},\"storageClassName\":\"longhorn-ubuntu-22\",\"volumeMode\":\"Block\"},\"status\":{}}]"
		},
		"labels": {
			"harvesterhci.io/creator": "harvester",
			"harvesterhci.io/os": "linux",
			"harvesterhci.io/vmName": "runner",
			"tag.harvesterhci.io/ssh-user": "ubuntu"
		},
		"name": "test-4",
		"namespace": "drew"
	},
	"spec": {
		"runStrategy": "RerunOnFailure",
		"template": {
			"metadata": {
				"annotations": {
					"harvesterhci.io/waitForLeaseInterfaceNames": "[\"nic-1\"]"
				},
				"labels": {
					"harvesterhci.io/creator": "terraform-provider-harvester",
					"harvesterhci.io/vmName": "test",
					"tag.harvesterhci.io/ssh-user": "ubuntu"
				}
			},
			"spec": {
				"affinity": {
					"nodeAffinity": {
						"requiredDuringSchedulingIgnoredDuringExecution": {
							"nodeSelectorTerms": [
								{
									"matchExpressions": [
										{
											"key": "network.harvesterhci.io/mgmt",
											"operator": "In",
											"values": [
												"true"
											]
										}
									]
								}
							]
						}
					},
					"podAntiAffinity": {
						"preferredDuringSchedulingIgnoredDuringExecution": [
							{
								"podAffinityTerm": {
									"labelSelector": {
										"matchExpressions": [
											{
												"key": "harvesterhci.io/creator",
												"operator": "Exists"
											}
										]
									},
									"topologyKey": "kubernetes.io/hostname"
								},
								"weight": 100
							}
						]
					}
				},
				"domain": {
					"cpu": {
						"cores": 1
					},
					"devices": {
						"disks": [
							{
								"bootOrder": 1,
								"disk": {
									"bus": "virtio"
								},
								"name": "rootdisk"
							},
							{
								"disk": {
									"bus": "virtio"
								},
								"name": "cloudinitdisk"
							}
						],
						"interfaces": [
							{
								"bridge": {},
								"macAddress": "1e:31:ed:d3:83:9c",
								"model": "virtio",
								"name": "nic-1"
							}
						]
					},
					"features": {
						"acpi": {},
						"smm": {
							"enabled": true
						}
					},
					"firmware": {
						"bootloader": {
							"efi": {
								"secureBoot": true
							}
						}
					},
					"machine": {
						"type": "q35"
					},
					"memory": {
						"guest": "1024Mi"
					},
					"resources": {
						"limits": {
							"cpu": "1",
							"memory": "2Gi"
						},
						"requests": {
							"cpu": "1",
							"memory": "2Gi"
						}
					}
				},
				"evictionStrategy": "LiveMigrate",
				"hostname": "test",
				"networks": [
					{
						"multus": {
							"networkName": "harvester-public/lab"
						},
						"name": "nic-1"
					}
				],
				"terminationGracePeriodSeconds": 120,
				"volumes": [
					{
						"name": "rootdisk",
						"persistentVolumeClaim": {
							"claimName": "packer-build"
						}
					},
					{
						"cloudInitNoCloud": {
							"networkDataSecretRef": {
								"name": "packer"
							},
							"secretRef": {
								"name": "packer"
							}
						},
						"name": "cloudinitdisk"
					}
				]
			}
		}
	}
 }
 `
