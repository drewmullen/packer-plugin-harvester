package harvester

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	harvester "github.com/drewmullen/harvester-go-sdk"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

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

func waitForVMStateDestroy(name string, namespace string, client harvester.APIClient, auth context.Context, timeout time.Duration, ui packersdk.Ui) error {
	startTime := time.Now()

	for {
		readReq := client.VirtualMachinesAPI.ReadNamespacedVirtualMachineInstance(auth, name, namespace)
		_, resp, err := readReq.Execute()

		if resp.StatusCode == http.StatusNotFound {
			ui.Say("VM has been destroyed")
			return nil
		}

		if err != nil {
			return err
		}

		if time.Since(startTime) >= timeout {
			return errors.New("timeout waiting for desired state")
		}

		ui.Say("Waiting for VM to be destroyed...")
		time.Sleep(10 * time.Second) // Adjust the polling interval as needed
	}
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

func waitForImageDownload(desiredState int32, name string, namespace string, client harvester.APIClient, auth context.Context, timeout time.Duration, ui packersdk.Ui) error {
	startTime := time.Now()
	for {
		readReq := client.ImagesAPI.ReadNamespacedVirtualMachineImage(auth, name, namespace)
		readImage, _, err := readReq.Execute()
		if err != nil {
			return err
		}

		progress := int32(0)
		// image.Status.Progress key doesnt appear until download progress starts
		if readImage.Status.HasProgress() {
			if *readImage.Status.Progress == desiredState {
				return nil
			} else {
				progress = *readImage.Status.Progress
			}
		}

		if time.Since(startTime) >= timeout {
			return errors.New("timeout waiting for desired state")
		}

		ui.Say(fmt.Sprintf("Download in progress... %v%%", progress))
		time.Sleep(5 * time.Second) // Adjust the polling interval as needed
	}
}
