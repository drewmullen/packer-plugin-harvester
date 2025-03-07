// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package harvester

import (
	"context"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"

	harvester "github.com/drewmullen/harvester-go-sdk"
)

const BuilderId = "harvester.builder"

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {

	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}

	return nil, warnings, nil
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
		&StepSourceBase{},
		&StepCreateVolume{},
		&StepCreateVM{},
		// TODO: on hold, cannot track status of exported VM
		// new(commonsteps.StepProvision),
		// &StepExportVMImage{},
	)

	// Set the value of the generated data that will become available to provisioners.
	// To share the data with post-processors, use the StateData in the artifact.
	/*
	state.Put("generated_data", map[string]interface{}{
		"GeneratedMockData": "mock-build-data",
	})*/

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
