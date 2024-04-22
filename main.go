package main

import (
	"fmt"
	"os"

	harvester "github.com/rptcloud/packer-plugin-harvester/builder/harvester"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder(plugin.DEFAULT_NAME, new(harvester.Builder))
	// pps.RegisterPostProcessor("import", new(digitaloceanPP.PostProcessor))
	// pps.RegisterDatasource("image", new(image.Datasource))
	// pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
