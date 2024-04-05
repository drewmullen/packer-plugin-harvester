// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-harvester/builder/harvester"
	harvesterData "github.com/hashicorp/packer-plugin-harvester/datasource/harvester"
	harvesterPP "github.com/hashicorp/packer-plugin-harvester/post-processor/harvester"
	harvesterProv "github.com/hashicorp/packer-plugin-harvester/provisioner/harvester"
	harvesterVersion "github.com/hashicorp/packer-plugin-harvester/version"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("my-builder", new(harvester.Builder))
	pps.RegisterProvisioner("my-provisioner", new(harvesterProv.Provisioner))
	pps.RegisterPostProcessor("my-post-processor", new(harvesterPP.PostProcessor))
	pps.RegisterDatasource("my-datasource", new(harvesterData.Datasource))
	pps.SetVersion(harvesterVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
