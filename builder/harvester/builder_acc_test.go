// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Run with: PACKER_ACC=1 go test -count 1 -v ./builder/harvester/builder_acc_test.go  -timeout=120m
package harvester

import (
	"fmt"
	//"io/ioutil"
	"os"
	"os/exec"
	//"regexp"
	"testing"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	//"github.com/hashicorp/packer-plugin-sdk/multistep"
	//"context"
	//harvester "github.com/drewmullen/harvester-go-sdk"
	//packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	
)

// testBuilderAccBasic_imageChecksum. url is provided, checksum is provided image exists. should create image
const testBuilderAccBasic_imageChecksum = `
source "harvester" "foo" {

	builder_source {
		name    = "testexist"
		url     = "http://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img"
		os_type = "ubuntu"
		checksum = "02cb10fb8aacc83a2765cb84f76f4a922895ffd8342cd077ed676b0315eaee4e515fec812ac99912d66e95fb977dbbbb402127cd22d344941e8b296e9ed87100"
	}

	builder_configuration {
	 name_prefix = "test-"
	}

	builder_target {
	}
  }
  build {
	sources = [
	  "source.harvester.foo",
	]
  }
  `

// testBuilderAccBasic_imageNoChecksum. url is provided, checksum is not provided image does not exist attempt create. should create new image
const testBuilderAccBasic_noChecksumImageFail = `
source "harvester" "foo" {

	builder_source {
	  name    = "testcreate"
	  url     = "http://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img"
	  os_type = "ubuntu"
	}
  
	builder_configuration {
	  name_prefix = "test-"
	}
  
	builder_target {}
  }
  build {
	sources = [
	  "source.harvester.foo",
	]
  }
`

// testBuilderAccBasic_noChecksumNoUrlImage. url is not provided, checksum is not provided image exists. test not in use

const testBuilderAccBasic_noChecksumNoUrlImage = `
source "harvester" "foo" {
	builder_source {
	  name    = "testexist"
	  os_type = "ubuntu"
	}

	builder_configuration {
	 name_prefix = "test-"
	}

	builder_target {}
  }
  build {
	sources = [
	  "source.harvester.foo",
	]
  }
`

// testBuilderAccBasic_diffChecksumImage. image with same name already exists in harvester, checksum is provided, url is provided but checksums are not the same.should exit
const testBuilderAccBasic_diffChecksumImage = `
source "harvester" "foo" {

	builder_source {
		name    = "testexist"
		url     = "http://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img"
		os_type = "ubuntu"
		checksum = "02cb10fb8aacc83a2765cb84f76f4a922895ffd8342cd077ed676b0315eaee4e515fe12ac9912d66e95fb977dbbbb402127cd22d344941e8b296e9ed87100"
	}

	builder_configuration {
	 name_prefix = "test-"
	}

	builder_target {}
  }
  build {
	sources = [
	  "source.harvester.foo",
	]
  }
`

// testBuilderAccBasic_urlNoChecksumImage. url provided no checksum image exists. 
const testBuilderAccBasic_urlNoChecksumImage = `
source "harvester" "foo" {
	builder_source {
		url     = "http://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img"
		name    = "testexist"
		os_type = "ubuntu"
	}

	builder_configuration {
	 name_prefix = "test-"
	}

	builder_target {}
  }
  build {
	sources = [
	  "source.harvester.foo",
	]
  }
`


// testBuilderAccBasic_imageChecksum. url is provided, checksum is provided image exists. should create vm

func TestAccBulder_imageDownloadWithChecksum(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name: "image_download_with_checksum_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: testBuilderAccBasic_imageChecksum,
		Type:     "harvester",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			logs, err := os.Open(logfile)
			if err != nil {
				return fmt.Errorf("Unable find %s", logfile)
			}
			defer logs.Close()

			if err != nil {
				return fmt.Errorf("Unable to read %s", logfile)
			}
			
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

// testBuilderAccBasic_imageNoChecksum. url is provided, checksum is not provided image does not exist attempt create
func TestAccBulder_imageDownloadWithoutChecksum(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name: "image_download_without_checksum_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: testBuilderAccBasic_noChecksumImageFail,
		Type:     "harvester",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			logs, err := os.Open(logfile)
			if err != nil {
				return fmt.Errorf("Unable find %s", logfile)
			}
			defer logs.Close()

			if err != nil {
				return fmt.Errorf("Unable to read %s", logfile)
			}
	
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

//test3. url is not provided, checksum is not provided image exists. should exit
func TestAccBuild_imageExistsDownload(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name: "imageExistsDownload",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: testBuilderAccBasic_noChecksumNoUrlImage,
		Type:     "harvester",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			logs, err := os.Open(logfile)
			if err != nil {
				return fmt.Errorf("Unable find %s", logfile)
			}
			defer logs.Close()

			if err != nil {
				return fmt.Errorf("Unable to read %s", logfile)
			}
			
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

// testBuilderAccBasic_diffChecksumImage. image with same name already exists in harvester, checksum is provided, url is provided but checksums differ should exit
func TestAccBuild_imageExistsWithDifferentChecksum(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name: "imageExistsWithDifferentChecksum",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: testBuilderAccBasic_diffChecksumImage,
		Type:     "harvester",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			logs, err := os.Open(logfile)
			if err != nil {
				return fmt.Errorf("Unable find %s", logfile)
			}
			defer logs.Close()

			if err != nil {
				return fmt.Errorf("Unable to read %s", logfile)
			}
	
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

// testBuilderAccBasic_urlNoChecksumImage. url provided no checksum image exists. test not in use
func TestAccBuild_imageExistsWithNoURLNoCheckSum(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name: "imageExistsWithNoURLNoCheckSum",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: testBuilderAccBasic_urlNoChecksumImage,
		Type:     "harvester",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			logs, err := os.Open(logfile)
			if err != nil {
				return fmt.Errorf("Unable find %s", logfile)
			}
			defer logs.Close()

			if err != nil {
				return fmt.Errorf("Unable to read %s", logfile)
			}

			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}
/*
	func TestAccBuild_testSetupUse(t *testing.T) {
		state := new(multistep.BasicStateBag)
		client := state.Get("client").(*harvester.APIClient)
		auth := state.Get("auth").(context.Context)
		checksum:="02cb10fb8aacc83a2765cb84f76f4a922895ffd8342cd077ed676b0315eaee4e515fe12ac9912d66e95fb977dbbbb402127cd22d344941e8b296e9ed87100"
		name    := "testexist"
		ui := state.Get("ui").(packersdk.Ui)

		testCase := &acctest.PluginTestCase{
			Name: "testSetUpUse",
			Setup: func() error {
			
				img,err :=checkImageExists(client, auth, name,"drew")
				if err != nil {
					return err
				}
				if (img!=harvester.HarvesterhciIoV1beta1VirtualMachineImage{}){
					waitForImageDownload(int32(100), name, "drew", *client, auth, 100, ui)
				}
				if *img.Spec.Checksum != checksum {
					return fmt.Errorf("Image already exists")
				}

				return nil
			},
			Teardown: func() error {
				client.VirtualMachinesAPI.DeleteNamespacedVirtualMachine(auth, name, "drew")
				return nil
			},
			Template: testBuilderAccBasic_urlNoChecksumImage,
			Type:     "harvester",
			Check: func(buildCommand *exec.Cmd, logfile string) error {
				if buildCommand.ProcessState != nil {
					if buildCommand.ProcessState.ExitCode() != 0 {
						return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
					}
				}
	
				logs, err := os.Open(logfile)
				if err != nil {
					return fmt.Errorf("Unable find %s", logfile)
				}
				defer logs.Close()
	
				//ogsBytes, err := ioutil.ReadAll(logs)
				if err != nil {
					return fmt.Errorf("Unable to read %s", logfile)
				}
				//logsString := string(logsBytes)
	
				/*buildGeneratedDataLog := "harvester.basic-example: build generated data: mock-build-data"
				if matched, _ := regexp.MatchString(buildGeneratedDataLog+".*", logsString); !matched {
					t.Fatalf("logs doesn't contain expected foo value %q", logsString)
				}
				return nil
			},
		}
	acctest.TestPlugin(t, testCase)
}
*/