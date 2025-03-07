  Include a short description about the provisioner. This is a good place
  to call out what the provisioner does, and any additional text that might
  be helpful to a user. See https://www.packer.io/docs/provisioner/null
-->

The harvester provisioner is used to provisioner Packer builds.


<!-- Provisioner Configuration Fields -->

**Required**

- `mock` (string) - The name of the mock string to display.


<!--
  Optional Configuration Fields

  Configuration options that are not required or have reasonable defaults
  should be listed under the optionals section. Defaults values should be
  noted in the description of the field
-->

**Optional**


<!--
  A basic example on the usage of the provisioner. Multiple examples
  can be provided to highlight various configurations.

-->
### Example Usage


```hcl
 source "null" "example" {
   communicator = "none"
 }

 build {
   source "null.example" {
     name = "jay"
   }

   provisioner "harvester" {
     mock = "mocking ${source.name}"
   }
 }
```
