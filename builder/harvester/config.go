//go:generate packer-sdc mapstructure-to-hcl2 -type Config,BuilderSource,BuilderConfiguration,BuilderTarget

package harvester

import (
	"os"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	HarvesterURL        string `mapstructure:"harvester_url"`
	HarvesterToken      string `mapstructure:"harvester_token"`
	HarvesterNamespace  string `mapstructure:"harvester_namespace"`

	BuilderSource        BuilderSource        `mapstructure:"builder_source"`
	BuilderConfiguration BuilderConfiguration `mapstructure:"builder_configuration"`
	BuilderTarget        BuilderTarget        `mapstructure:"builder_target"`
}

type BuilderSource struct {
	Name      string `mapstructure:"name"`
	OSType    string `mapstructure:"os_type"`
	ImageType string `mapstructure:"image_type"`

	URL         string `mapstructure:"url" required:"false"`
	DisplayName string `mapstructure:"display_name" required:"false"`
	Checksum    string `mapstructure:"checksum" required:"false"`
	Cleanup     bool   `mapstructure:"cleanup" required:"false"`
}

type BuilderConfiguration struct {
	// default to HarvesterNamespace
	Namespace string `mapstructure:"namespace" required:"false"`
	// default to "packer-"
	NamePrefix string `mapstructure:"name_prefix" required:"false"`
	// default 1
	CPU int64 `mapstructure:"cpu" required:"false"`
	// default 2Gi
	Memory string `mapstructure:"memory" required:"false"`
	// PreventBuilderImageCleanup bool `mapstructure:"prevent_builder_image_cleanup" required:"false"`
	NetworkNamespace string `mapstructure:"network_namespace"`
	Network string `mapstructure:"network"`
}

type BuilderTarget struct {
	// default to HarvesterNamespace
	Namespace   string `mapstructure:"namespace" required:"false"`
	DisplayName string `mapstructure:"display_name" required:"false"`
	VolumeSize  string `mapstructure:"volume_size" required:"false"`
}

func (c *Config) Prepare(raws ...interface{}) (generatedVars []string, err error) {
	err = config.Decode(c, &config.DecodeOpts{
		PluginType:  "packer.builder.harvester",
		Interpolate: true,
	}, raws...)
	if err != nil {
		return nil, err
	}

	if c.HarvesterURL == "" {
		c.HarvesterURL = os.Getenv("HARVESTER_URL")
	}
	if c.HarvesterToken == "" {
		c.HarvesterToken = os.Getenv("HARVESTER_TOKEN")
	}
	if c.HarvesterNamespace == "" {
		c.HarvesterNamespace = os.Getenv("HARVESTER_NAMESPACE")
	}

	if c.BuilderConfiguration.Namespace == "" {
		c.BuilderConfiguration.Namespace = c.HarvesterNamespace
	}

	if c.BuilderConfiguration.NamePrefix == "" {
		c.BuilderConfiguration.NamePrefix = "packer-"
	}

	if c.BuilderConfiguration.CPU == 0 {
		c.BuilderConfiguration.CPU = 1
	}

	if c.BuilderConfiguration.Memory == "" {
		c.BuilderConfiguration.Memory = "2Gi"
	}

	if c.BuilderTarget.VolumeSize == "" {
		c.BuilderTarget.VolumeSize = "100Gi"
	}

	if c.BuilderConfiguration.NamePrefix == "" {
		c.BuilderConfiguration.NamePrefix = "packer-"
	}

	if c.BuilderConfiguration.NetworkNamespace == "" {
		c.BuilderConfiguration.NetworkNamespace = "harvester-public"
	}

	// Return the placeholder for the generated data that will become available to provisioners and post-processors.
	// If the builder doesn't generate any data, just return an empty slice of string: []string{}
	// buildGeneratedData := []string{"GeneratedMockData"}
	return []string{}, nil
}
