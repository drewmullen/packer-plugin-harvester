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

const testBuilderHCL2Basic = `
source "harvester" "foo" {
	builder_source {
	  name    = "drewbuntu"
	  url     = "http://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img"
	  os_type = "ubuntu"
	  checksum = "02cb10fb8aacc83a2765cb84f76f4a922895ffd8342cd077ed676b0315eaee4e515fec812ac99912d66e95fb977dbbbb402127cd22d344941e8b296e9ed87100"
	}
  
	builder_configuration {
	 name_prefix = "drew-" 
	}
  
	builder_target {}
  }
  build {
	sources = [
	  "source.harvester.foo",
	]
  }
  `

// Run with: PACKER_ACC=1 go test -count 1 -v ./builder/harvester/builder_acc_test.go  -timeout=120m
func TestAccHarvesterBuilder(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name: "harvester_builder_basic_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: testBuilderHCL2Basic,
		Type:     "harvester-img",
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

			buildGeneratedDataLog := "harvester-img.basic-example: build generated data: mock-build-data"
			if matched, _ := regexp.MatchString(buildGeneratedDataLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected foo value %q", logsString)
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}
