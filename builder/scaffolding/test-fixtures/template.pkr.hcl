# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

source "harvester-my-builder" "basic-example" {
  mock = "mock-config"
}

build {
  sources = [
    "source.harvester-my-builder.basic-example"
  ]

  provisioner "shell-local" {
    inline = [
      "echo build generated data: ${build.GeneratedMockData}",
    ]
  }
}
