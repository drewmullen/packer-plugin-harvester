// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
// image download tests
// 1. url is provided, checksum is provided
// 2. url is provided, checksum is not provided
// 3. url is not provided, checksum is not provided
// 4. image with same name already exists in harvester, checksum is provided, url is provided but checksums differ, expect failure

// TestAccBulder_imageDownloadWithChecksum
// TestAccBulder_imageDownloadWithoutChecksum
// TestAccBuild_imageExistsNoDownload
// TestAccBuild_imageExistsWithDifferentChecksum
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

const test1 = `
source "harvester" "foo" {

	harvester_url="https://rancher.danquack.dev/k8s/clusters/c-m-mnff6d8p"
	harvester_token="token-sqqf5:x72vh9wr8pzpjmbstcwpwsmk87hlqb6dqzft28nsb72ztlsphgwdlw"
	harvester_namespace="drew"
	
	builder_source {
	  name    = "drewbuntu"
	  url     = "http://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img"
	  os_type = "ubuntu"
	  checksum = "02cb10fb8aacc83a2765cb84f76f4a922895ffd8342cd077ed676b0315eaee4e515fec812ac99912d66e95fb977dbbbb402127cd22d344941e8b296e9ed87100"
	}
  
	builder_configuration {
	 name_prefix = "drew-"
	 namespace= "drew"
	}
  
	builder_target {}
  }
  build {
	sources = [
	  "source.harvester.foo",
	]
  }
  `
const test2=`
source "harvester" "foo" {

	harvester_url="https://rancher.danquack.dev/k8s/clusters/c-m-mnff6d8p"
	harvester_token="token-sqqf5:x72vh9wr8pzpjmbstcwpwsmk87hlqb6dqzft28nsb72ztlsphgwdlw"
	harvester_namespace="drew"
	
	builder_source {
	  name    = "drewbuntu"
	  url     = "http://cloud-images.ubuntu.com/releases/focal/release/ubuntu-20.04-server-cloudimg-amd64.img"
	  os_type = "ubuntu"
	}
  
	builder_configuration {
	 name_prefix = "drew-"
	 namespace= "drew"
	}
  
	builder_target {}
  }
  build {
	sources = [
	  "source.harvester.foo",
	]
  }
`

const test3=`
source "harvester" "foo" {

	harvester_url="https://rancher.danquack.dev/k8s/clusters/c-m-mnff6d8p"
	harvester_token="token-sqqf5:x72vh9wr8pzpjmbstcwpwsmk87hlqb6dqzft28nsb72ztlsphgwdlw"
	harvester_namespace="drew"
	
	builder_source {
	  name    = "drewbuntu"
	  os_type = "ubuntu"
	  checksum = "02cb10fb8aacc83a2765cb84f76f4a922895ffd8342cd077ed676b0315eaee4e515fec812ac99912d66e95fb977dbbbb402127cd22d344941e8b296e9ed87100"
	}
  
	builder_configuration {
	 name_prefix = "drew-"
	 namespace= "drew"
	}
  
	builder_target {}
  }
  build {
	sources = [
	  "source.harvester.foo",
	]
  }
`
const test4=`
source "harvester" "foo" {

	harvester_url="https://rancher.danquack.dev/k8s/clusters/c-m-mnff6d8p"
	harvester_token="token-sqqf5:x72vh9wr8pzpjmbstcwpwsmk87hlqb6dqzft28nsb72ztlsphgwdlw"
	harvester_namespace="drew"
	
	builder_source {
	  name    = "drewbuntu"
	  os_type = "ubuntu"
	}
  
	builder_configuration {
	 name_prefix = "drew-"
	 namespace= "drew"
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
func TestAccBulder_imageDownloadWithChecksum(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name: "harvester_builder_basic_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: test1,
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

func TestAccBulder_imageDownloadWithoutChecksum(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name: "harvester_builder_basic_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: test2,
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

func TestAccBuild_imageExistsNoDownload(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name: "harvester_builder_basic_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: test3,
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

func TestAccBuild_imageExistsWithDifferentChecksum(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name: "harvester_builder_basic_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: test4,
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