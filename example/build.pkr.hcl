# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    harvester = {
      version = ">=v0.1.0"
      source  = "github.com/rptcloud/harvester"
    }
  }
}

source "harvester-my-builder" "foo-example" {
  mock = local.foo
}

source "harvester-my-builder" "bar-example" {
  mock = local.bar
}

build {
  sources = [
    "source.harvester-my-builder.foo-example",
  ]

  source "source.harvester-my-builder.bar-example" {
    name = "bar"
  }

  provisioner "harvester-my-provisioner" {
    only = ["harvester-my-builder.foo-example"]
    mock = "foo: ${local.foo}"
  }

  provisioner "harvester-my-provisioner" {
    only = ["harvester-my-builder.bar"]
    mock = "bar: ${local.bar}"
  }

  post-processor "harvester-my-post-processor" {
    mock = "post-processor mock-config"
  }
}
