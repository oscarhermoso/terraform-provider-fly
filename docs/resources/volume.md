---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "fly_volume Resource - terraform-provider-fly"
subcategory: ""
description: |-
  Fly volume resource
---

# fly_volume (Resource)

Fly volume resource

## Example Usage

```terraform
resource "fly_volume" "exampleApp" {
  name   = "exampleVolume"
  app    = "hellofromterraform"
  size   = 10
  region = "ewr"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app` (String) Name of app to attach to
- `name` (String) name
- `region` (String) region
- `size` (Number) Size of volume in GB

### Optional

- `id` (String) ID of volume
- `internalid` (String) Internal ID

## Import

Import is supported using the following syntax:

```shell
terraform import fly_volume.exampleApp <app_id>,<volume_internal_id>
```
