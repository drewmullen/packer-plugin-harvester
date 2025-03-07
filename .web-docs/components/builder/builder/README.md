  Include a short description about the builder. This is a good place
  to call out what the builder does, and any requirements for the given
  builder environment. See https://www.packer.io/docs/builder/null
-->

The harvester builder is used to create endless Packer plugins using
a consistent plugin structure.


<!-- Builder Configuration Fields -->

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
  A basic example on the usage of the builder. Multiple examples
  can be provided to highlight various build configurations.

-->
### Example Usage


```hcl
 source "harvester" "example" {
   mock = "bird"
 }

 build {
   sources = ["source.harvester.example"]
 }
```
