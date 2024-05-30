// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package harvester

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

//go:embed test-fixtures/template.pkr.hcl
var testDatasourceHCL2Basic string

// image download tests
// 1. url is provided, checksum is provided
// 2. url is provided, checksum is not provided
// 3. url is not provided, checksum is not provided
// 4. image with same name already exists in harvester, checksum is provided, url is provided but checksums differ, expect failure

// TestAccBulder_imageDownloadWithChecksum
// TestAccBulder_imageDownloadWithoutChecksum
// TestAccBuild_imageExistsNoDownload
// TestAccBuild_imageExistsWithDifferentChecksum

// Run with: PACKER_ACC=1 go test -count 1 -v ./datasource/harvester/data_acc_test.go  -timeout=120m

func TestAccBuilder_imageDownloadWithChecksum(t *testing.T){
	testCase := &acctest.PluginTestCase{
		Name: "template",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: `source "harvester" "basic-example" {
			// given via environment variables
			 harvester_url = "https://rancher.danquack.dev/k8s/clusters/c-m-84qqjh2s"
			 harvester_namespace = "drew"
			 harvester_token = "token-hnkfp:v2z49q24txp8b8m8p9swljqrfhz8864ss24nd4jpd6qx6wwwp6r5sc"
			 
		  }
		  
		  build {
			sources = [
			  "source.harvester"
			]
		  
			provisioner "shell-local" {
			  inline = [
				"echo build generated data: ${build.GeneratedMockData}",
			  ]
			}
		  }
		  
		  
		  build {
			sources = [
			  "source.harvester.basic-example"
			]
		  
			provisioner "shell-local" {
			inline = [
				 "echo build generated data: ${build.GeneratedMockData}",
			   ]
		   }
		  }
		  `,
		Type:     "harvester-my-datasource",
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

			logsBytes, err := ioutil.ReadAll(logs)
			if err != nil {
				return fmt.Errorf("Unable to read %s", logfile)
			}
			logsString := string(logsBytes)

			fooLog := "null.basic-example: foo: foo-value"
			barLog := "null.basic-example: bar: bar-value"

			if matched, _ := regexp.MatchString(fooLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected foo value %q", logsString)
			}
			if matched, _ := regexp.MatchString(barLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected bar value %q", logsString)
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}
/*func TestAccHarvesterDatasource(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name: "harvester_datasource_basic_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: testDatasourceHCL2Basic,
		Type:     "harvester-my-datasource",
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

			logsBytes, err := ioutil.ReadAll(logs)
			if err != nil {
				return fmt.Errorf("Unable to read %s", logfile)
			}
			logsString := string(logsBytes)

			fooLog := "null.basic-example: foo: foo-value"
			barLog := "null.basic-example: bar: bar-value"

			if matched, _ := regexp.MatchString(fooLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected foo value %q", logsString)
			}
			if matched, _ := regexp.MatchString(barLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected bar value %q", logsString)
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}
*/