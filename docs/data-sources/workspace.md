---
page_title: "webex_workspace Data Source"
subcategory: ""
description: |-
  Looks up a Webex Workspace by display name.
---

# webex_workspace (Data Source)

Looks up a Webex Workspace by exact display name. Returns the workspace's full configuration including calling, calendar, and device hosted meetings settings.

## Example Usage

```hcl
data "webex_workspace" "boardroom" {
  display_name = "5th Floor Boardroom"
}

output "boardroom_id" {
  value = data.webex_workspace.boardroom.id
}
```

## Schema

### Required

- `display_name` (String) — Display name of the workspace to look up (exact match).

### Read-Only

- `id` (String) — Workspace ID.
- `org_id` (String) — Organization ID.
- `location_id` (String) — Location ID.
- `floor_id` (String) — Floor ID.
- `capacity` (Number) — Room capacity.
- `type` (String) — Workspace type.
- `sip_address` (String) — SIP address.
- `created` (String) — Creation timestamp (ISO 8601).
- `notes` (String) — Notes.
- `hotdesking_status` (String) — Hot desking status.
- `supported_devices` (String) — Supported device type.
- `device_platform` (String) — Device platform.
- `indoor_navigation_url` (String) — Indoor navigation map URL.

### Blocks

#### `calling`

- `type` (String) — Calling type.
- `phone_number` (String) — Phone number.
- `extension` (String) — Extension.
- `location_id` (String) — Calling location ID.
- `licenses` (List of String) — Webex Calling licenses.

#### `calendar`

- `type` (String) — Calendar type.
- `email_address` (String) — Calendar email address.
- `resource_group_id` (String) — Calendar resource group ID.

#### `device_hosted_meetings`

- `enabled` (Boolean) — Whether device hosted meetings are enabled.
- `site_url` (String) — Webex site URL.
