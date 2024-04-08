package main

import (
	"fmt"
	"os"

	img "github.com/rptcloud/packer-plugin-harvester/builder/harvester/img"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("img", new(img.Builder))
	// pps.RegisterPostProcessor("import", new(digitaloceanPP.PostProcessor))
	// pps.RegisterDatasource("image", new(image.Datasource))
	// pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
