---
page_title: "webex_location Resource"
subcategory: ""
description: |-
  Manages a Webex Location — a named physical site with address, timezone, and coordinates.
---

# webex_location

Manages a Webex Location. Locations represent physical sites (offices, buildings) and provide timezone and address context for workspaces and devices within them.

## Example Usage

```hcl
resource "webex_location" "oslo" {
  name      = "Oslo HQ"
  address1  = "Stortingsgata 6"
  city      = "Oslo"
  post_code = "0161"
  country   = "NO"
  time_zone = "Europe/Oslo"
  latitude  = "59.913868"
  longitude = "10.752245"
}
```

## Schema

### Required

- `name` (String) — The name of the location.

### Optional

- `address1` (String) — The primary street address.
- `address2` (String) — Additional address line.
- `city` (String) — The city.
- `state` (String) — The state or region.
- `post_code` (String) — The postal code.
- `country` (String) — The country code (ISO 3166-1).
- `time_zone` (String) — The time zone (e.g. `Europe/Oslo`).
- `latitude` (String) — The latitude.
- `longitude` (String) — The longitude.

### Read-Only

- `id` (String) — Unique identifier for the location.
- `org_id` (String) — Organization ID.

## Import

Locations can be imported using their ID:

```shell
terraform import webex_location.example Y2lzY29zcGFyazovL3...
```
