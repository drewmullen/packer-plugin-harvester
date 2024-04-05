  Include a short description about the data source. This is a good place
  to call out what the data source does, and any requirements for the given
  data source environment. See https://www.packer.io/docs/data-source/amazon-ami
-->

The harvester data source is used to create endless Packer plugins using
a consistent plugin structure.


<!-- Data source Configuration Fields -->

**Required**

- `mock` (string) - The name of the mock to use for the Harvester API.


<!--
  Optional Configuration Fields

  Configuration options that are not required or have reasonable defaults
  should be listed under the optionals section. Defaults values should be
  noted in the description of the field
-->

**Optional**

- `mock_api_url` (string) - The Harvester API endpoint to connect to.
  Defaults to https://example.com



<!--
  A basic example on the usage of the data source. Multiple examples
  can be provided to highlight various build configurations.

-->

### OutPut

- `foo` (string) - The Harvester output foo value.
- `bar` (string) - The Harvester output bar value.

<!--
  A basic example on the usage of the data source. Multiple examples
  can be provided to highlight various build configurations.

-->

### Example Usage


```hcl
data "harvester" "example" {
   mock = "bird"
 }
 source "harvester" "example" {
   mock = data.harvester.example.foo
 }

 build {
   sources = ["source.harvester.example"]
 }
```
