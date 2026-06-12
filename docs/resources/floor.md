---
page_title: "webex_floor Resource"
subcategory: ""
description: |-
  Manages a floor within a Webex Location.
---

# webex_floor

Manages a floor within a Webex Location. Floors help organize workspaces within a building.

## Example Usage

```hcl
resource "webex_floor" "ground" {
  location_id  = webex_location.oslo.id
  floor_number = 1
  display_name = "Ground Floor"
}

resource "webex_floor" "basement" {
  location_id  = webex_location.oslo.id
  floor_number = -1
  display_name = "Basement"
}
```

## Schema

### Required

- `location_id` (String) — The location this floor belongs to. Changing this forces a new resource.
- `floor_number` (Number) — The floor number. Can be negative (e.g. `-1` for basement).

### Optional

- `display_name` (String) — Display name for the floor.

### Read-Only

- `id` (String) — Unique identifier for the floor.

## Import

Floors can be imported using the format `location_id/floor_id`:

```shell
terraform import webex_floor.example Y2lzY29zcGFyazovL3.../Z2xv...
```
