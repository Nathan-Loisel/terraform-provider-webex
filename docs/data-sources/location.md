---
page_title: "webex_location Data Source"
subcategory: ""
description: |-
  Looks up a Webex Location by name.
---

# webex_location (Data Source)

Looks up a Webex Location by exact name. Returns the location's ID, address, timezone, and coordinates.

## Example Usage

```hcl
data "webex_location" "oslo" {
  name = "Protector Forsikring - Oslo"
}

resource "webex_workspace" "room" {
  display_name = "Meeting Room"
  location_id  = data.webex_location.oslo.id
}
```

## Schema

### Required

- `name` (String) — The exact name of the location to look up.

### Read-Only

- `id` (String) — Unique identifier for the location.
- `org_id` (String) — Organization ID.
- `address` (String) — Primary street address.
- `address2` (String) — Additional address line.
- `city` (String) — City.
- `state` (String) — State or region.
- `post_code` (String) — Postal code.
- `country` (String) — Country code (ISO 3166-1).
- `time_zone` (String) — Time zone.
- `latitude` (String) — Latitude.
- `longitude` (String) — Longitude.
