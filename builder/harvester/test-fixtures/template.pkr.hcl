# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

source "harvester" "basic-example" {
  // given via environment variables
  // harvester_url = ""
  // harvester_namespace = ""
  // harvester_token = ""
}

build {
  sources = [
    "source.harvester"
  ]

  // provisioner "shell-local" {
  //   inline = [
  //     "echo build generated data: ${build.GeneratedMockData}",
  //   ]
  // }
}