// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package img

import (
	"context"
	"os"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"

	harvester "github.com/drewmullen/harvester-go-sdk"
)

const BuilderId = "harvester.builder"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	HarvesterURL        string `mapstructure:"harvester_url"`
	HarvesterToken      string `mapstructure:"harvester_token"`
	HarvesterNamespace  string `mapstructure:"harvester_namespace"`
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) (generatedVars []string, warnings []string, err error) {
	err = config.Decode(&b.config, &config.DecodeOpts{
		PluginType:  "packer.builder.harvester",
		Interpolate: true,
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	if b.config.HarvesterURL == "" {
		b.config.HarvesterURL = os.Getenv("HARVESTER_URL")
	}
	if b.config.HarvesterToken == "" {
		b.config.HarvesterToken = os.Getenv("HARVESTER_TOKEN")
	}
	if b.config.HarvesterNamespace == "" {
		b.config.HarvesterNamespace = os.Getenv("HARVESTER_NAMESPACE")
	}

	// Return the placeholder for the generated data that will become available to provisioners and post-processors.
	// If the builder doesn't generate any data, just return an empty slice of string: []string{}
	// buildGeneratedData := []string{"GeneratedMockData"}
	return []string{}, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	steps := []multistep.Step{}

	configuration := &harvester.Configuration{
		DefaultHeader: make(map[string]string),
		UserAgent:     "OpenAPI-Generator/1.0.0/go",
		Debug:         false,
		Servers: harvester.ServerConfigurations{
			{
				URL:         b.config.HarvesterURL,
				Description: "Harvester API Server",
			},
		},
	}
	auth := context.WithValue(context.Background(), harvester.ContextAccessToken, b.config.HarvesterToken)
	client := harvester.NewAPIClient(configuration)

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("config", &b.config)
	state.Put("client", client)
	state.Put("auth", auth)

	steps = append(steps,
		&StepCreateVM{},
		// new(commonsteps.StepProvision),
	)

	// Set the value of the generated data that will become available to provisioners.
	// To share the data with post-processors, use the StateData in the artifact.
	state.Put("generated_data", map[string]interface{}{
		"GeneratedMockData": "mock-build-data",
	})

	// Run!
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if err, ok := state.GetOk("error"); ok {
		return nil, err.(error)
	}

	artifact := &Artifact{
		// Add the builder generated data to the artifact StateData so that post-processors
		// can access them.
		StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
	}
	return artifact, nil
}
