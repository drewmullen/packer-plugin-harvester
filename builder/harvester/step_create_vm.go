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
type StepCreateVM struct {
	Name string
}

// Run should execute the purpose of this step
func (s *StepCreateVM) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	client := state.Get("client").(*harvester.APIClient)
	auth := state.Get("auth").(context.Context)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	// Prepare values for temporary build VM
	if c.BuilderConfiguration.Namespace == "" {
		vmObject.Metadata.Namespace = toStringPtr(c.HarvesterNamespace)
	} else {
		vmObject.Metadata.Namespace = toStringPtr(c.BuilderConfiguration.Namespace)
	}

	if c.BuilderConfiguration.NamePrefix == "" {
		vmObject.Metadata.GenerateName = toStringPtr("packer-")
	} else {
		vmObject.Metadata.GenerateName = &c.BuilderConfiguration.NamePrefix
	}

	if c.BuilderConfiguration.CPU != 0 {
		vmObject.Spec.Template.Spec.Domain.Cpu.Cores = &c.BuilderConfiguration.CPU
	}

	if c.BuilderConfiguration.Memory == "" {
		vmObject.Spec.Template.Spec.Domain.Resources.Requests["memory"] = "2Gi"
	} else {
		vmObject.Spec.Template.Spec.Domain.Resources.Requests["memory"] = c.BuilderConfiguration.Memory
	}

	req := client.VirtualMachinesAPI.CreateNamespacedVirtualMachine(auth, c.HarvesterNamespace)

	req = req.KubevirtIoApiCoreV1VirtualMachine(*vmObject)
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
func (s *StepCreateVM) Cleanup(_ multistep.StateBag) {
	// Nothing to clean
}

func waitForVMState(desiredState string, name string, namespace string, client harvester.APIClient, auth context.Context, timeout time.Duration, ui packersdk.Ui) error {
	startTime := time.Now()

	for {
		readReq := client.VirtualMachinesAPI.ReadNamespacedVirtualMachineInstance(auth, name, namespace)
		currentState, _, err := readReq.Execute()
		if err != nil {
			return err
		}

		// TODO: handle failure states
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

var vmObject = &harvester.KubevirtIoApiCoreV1VirtualMachine{
	ApiVersion: &ApiVersionKubevirt,
	Kind:       &KindVirtualMachine,
	Metadata: &harvester.K8sIoV1ObjectMeta{
		Annotations: &map[string]string{
			"harvesterhci.io/vmRunStrategy":            "RerunOnFailure",
			"kubevirt.io/latest-observed-api-version":  "v1",
			"kubevirt.io/storage-observed-api-version": "v1alpha3",
			"network.harvesterhci.io/ips":              "[]",
			"harvesterhci.io/volumeClaimTemplates":     "[{\"metadata\":{\"name\":\"packer-build\",\"creationTimestamp\":null,\"annotations\":{\"harvesterhci.io/imageId\":\"drew/drewbuntu\",\"terraform-provider-harvester-auto-delete\":\"true\"}},\"spec\":{\"accessModes\":[\"ReadWriteMany\"],\"resources\":{\"requests\":{\"storage\":\"100Gi\"}},\"storageClassName\":\"longhorn-ubuntu-22\",\"volumeMode\":\"Block\"},\"status\":{}}]",
		},
		Labels: &map[string]string{
			"harvesterhci.io/creator":      "harvester",
			"harvesterhci.io/os":           "linux",
			"harvesterhci.io/vmName":       "runner",
			"tag.harvesterhci.io/ssh-user": "ubuntu",
		},
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
						Cores: toInt64Ptr(1),
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
						Requests: map[string]string{},
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
							ClaimName: "packer-build",
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

// var vmStr string = `{
// 	"apiVersion": "kubevirt.io/v1",
// 	"kind": "VirtualMachine",
// 	"metadata": {
// 		"annotations": {
// 			"harvesterhci.io/vmRunStrategy": "RerunOnFailure",
// 			"kubevirt.io/latest-observed-api-version": "v1",
// 			"kubevirt.io/storage-observed-api-version": "v1alpha3",
// 			"network.harvesterhci.io/ips": "[]",
// 			"harvesterhci.io/volumeClaimTemplates": "[{\"metadata\":{\"name\":\"packer-build\",\"creationTimestamp\":null,\"annotations\":{\"harvesterhci.io/imageId\":\"drew/drewbuntu\",\"terraform-provider-harvester-auto-delete\":\"true\"}},\"spec\":{\"accessModes\":[\"ReadWriteMany\"],\"resources\":{\"requests\":{\"storage\":\"100Gi\"}},\"storageClassName\":\"longhorn-ubuntu-22\",\"volumeMode\":\"Block\"},\"status\":{}}]"
// 		},
// 		"labels": {
// 			"harvesterhci.io/creator": "harvester",
// 			"harvesterhci.io/os": "linux",
// 			"harvesterhci.io/vmName": "runner",
// 			"tag.harvesterhci.io/ssh-user": "ubuntu"
// 		},
// 		"name": "test-4",
// 		"namespace": "drew"
// 	},
// 	"spec": {
// 		"runStrategy": "RerunOnFailure",
// 		"template": {
// 			"metadata": {
// 				"annotations": {
// 					"harvesterhci.io/waitForLeaseInterfaceNames": "[\"nic-1\"]"
// 				},
// 				"labels": {
// 					"harvesterhci.io/creator": "terraform-provider-harvester",
// 					"harvesterhci.io/vmName": "test",
// 					"tag.harvesterhci.io/ssh-user": "ubuntu"
// 				}
// 			},
// 			"spec": {
// 				"affinity": {
// 					"nodeAffinity": {
// 						"requiredDuringSchedulingIgnoredDuringExecution": {
// 							"nodeSelectorTerms": [
// 								{
// 									"matchExpressions": [
// 										{
// 											"key": "network.harvesterhci.io/mgmt",
// 											"operator": "In",
// 											"values": [
// 												"true"
// 											]
// 										}
// 									]
// 								}
// 							]
// 						}
// 					},
// 					"podAntiAffinity": {
// 						"preferredDuringSchedulingIgnoredDuringExecution": [
// 							{
// 								"podAffinityTerm": {
// 									"labelSelector": {
// 										"matchExpressions": [
// 											{
// 												"key": "harvesterhci.io/creator",
// 												"operator": "Exists"
// 											}
// 										]
// 									},
// 									"topologyKey": "kubernetes.io/hostname"
// 								},
// 								"weight": 100
// 							}
// 						]
// 					}
// 				},
// 				"domain": {
// 					"cpu": {
// 						"cores": 1
// 					},
// 					"devices": {
// 						"disks": [
// 							{
// 								"bootOrder": 1,
// 								"disk": {
// 									"bus": "virtio"
// 								},
// 								"name": "rootdisk"
// 							},
// 							{
// 								"disk": {
// 									"bus": "virtio"
// 								},
// 								"name": "cloudinitdisk"
// 							}
// 						],
// 						"interfaces": [
// 							{
// 								"bridge": {},
// 								"macAddress": "1e:31:ed:d3:83:9c",
// 								"model": "virtio",
// 								"name": "nic-1"
// 							}
// 						]
// 					},
// 					"features": {
// 						"acpi": {},
// 						"smm": {
// 							"enabled": true
// 						}
// 					},
// 					"firmware": {
// 						"bootloader": {
// 							"efi": {
// 								"secureBoot": true
// 							}
// 						}
// 					},
// 					"machine": {
// 						"type": "q35"
// 					},
// 					"memory": {
// 						"guest": "1024Mi"
// 					},
// 					"resources": {
// 						"limits": {
// 							"cpu": "1",
// 							"memory": "2Gi"
// 						},
// 						"requests": {
// 							"cpu": "1",
// 							"memory": "2Gi"
// 						}
// 					}
// 				},
// 				"evictionStrategy": "LiveMigrate",
// 				"hostname": "test",
// 				"networks": [
// 					{
// 						"multus": {
// 							"networkName": "harvester-public/lab"
// 						},
// 						"name": "nic-1"
// 					}
// 				],
// 				"terminationGracePeriodSeconds": 120,
// 				"volumes": [
// 					{
// 						"name": "rootdisk",
// 						"persistentVolumeClaim": {
// 							"claimName": "packer-build"
// 						}
// 					},
// 					{
// 						"cloudInitNoCloud": {
// 							"networkDataSecretRef": {
// 								"name": "packer"
// 							},
// 							"secretRef": {
// 								"name": "packer"
// 							}
// 						},
// 						"name": "cloudinitdisk"
// 					}
// 				]
// 			}
// 		}
// 	}
//  }
//  `
