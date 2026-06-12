---
page_title: "webex_devices Data Source"
subcategory: ""
description: |-
  Lists Webex devices, optionally filtered by workspace.
---

# webex_devices (Data Source)

Lists Webex devices registered in your organization. Can be filtered by workspace to find devices in a specific room.

## Example Usage

### List All Devices

```hcl
data "webex_devices" "all" {}

output "device_count" {
  value = length(data.webex_devices.all.devices)
}
```

### Devices in a Specific Workspace

```hcl
data "webex_devices" "boardroom" {
  workspace_id = webex_workspace.boardroom.id
}

output "boardroom_devices" {
  value = [for d in data.webex_devices.boardroom.devices : d.display_name]
}
```

## Schema

### Optional

- `workspace_id` (String) — Filter devices by workspace ID.

### Read-Only

- `devices` (List of Object) — List of devices. Each device contains:
  - `id` (String) — Device ID.
  - `display_name` (String) — Device display name.
  - `workspace_id` (String) — Workspace ID.
  - `person_id` (String) — Person ID if assigned to a person.
  - `org_id` (String) — Organization ID.
  - `product` (String) — Product name.
  - `type` (String) — Device type.
  - `serial` (String) — Serial number.
  - `connection_status` (String) — Connection status.
  - `ip` (String) — IP address.
  - `mac` (String) — MAC address.
  - `software` (String) — Software version.
  - `primary_sip_url` (String) — Primary SIP URL.
  - `upgrade_channel` (String) — Upgrade channel.
  - `created` (String) — Creation timestamp (ISO 8601).
  - `first_seen` (String) — First seen timestamp (ISO 8601).
  - `last_seen` (String) — Last seen timestamp (ISO 8601).
  - `location_id` (String) — Location ID.
  - `managed_by` (String) — Entity managing the device.
  - `device_platform` (String) — Device platform.
  - `tags` (List of String) — Tags assigned to the device.
