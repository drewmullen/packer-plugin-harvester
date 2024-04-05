# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

locals {
  foo = data.harvester-my-datasource.mock-data.foo
  bar = data.harvester-my-datasource.mock-data.bar
}