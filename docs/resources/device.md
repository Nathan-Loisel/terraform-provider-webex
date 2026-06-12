---
page_title: "webex_device Resource"
subcategory: ""
description: |-
  Manages a Webex Device — creates a phone by MAC address in a workspace or for a person.
---

# webex_device

Manages a Webex Device. Creates a phone by its MAC address and assigns it to a workspace or a person.

Most attributes are read-only since they come from the physical device. Only `tags` can be updated in-place — all other changes force replacement.

## Example Usage

### Device in a Workspace

```hcl
resource "webex_device" "room_codec" {
  mac          = "AA:BB:CC:DD:EE:FF"
  model        = "Cisco Room Kit"
  workspace_id = webex_workspace.boardroom.id
  tags         = ["boardroom", "managed", "5th-floor"]
}
```

### Device for a Person

```hcl
resource "webex_device" "desk_phone" {
  mac       = "11:22:33:44:55:66"
  model     = "Cisco 8845"
  person_id = "Y2lzY29zcGFyazovL3..."
}
```

## Schema

### Required

- `mac` (String) — MAC address of the device. Changing this forces a new resource.
- `model` (String) — Model of the device. Changing this forces a new resource.

### Optional

- `workspace_id` (String) — The workspace to assign the device to. Mutually exclusive with `person_id`. Changing this forces a new resource.
- `person_id` (String) — The person to assign the device to. Mutually exclusive with `workspace_id`. Changing this forces a new resource.
- `password` (String, Sensitive) — SIP password for third-party devices. Changing this forces a new resource.
- `tags` (List of String) — Tags assigned to the device. This is the only attribute that can be updated in-place.

### Read-Only

- `id` (String) — Unique identifier for the device.
- `display_name` (String) — Device display name.
- `org_id` (String) — Organization ID.
- `product` (String) — Product name.
- `type` (String) — Device type: `roomdesk`, `phone`, `accessory`, `webexgo`, `unknown`.
- `serial` (String) — Serial number.
- `software` (String) — Software version.
- `ip` (String) — Current IP address.
- `primary_sip_url` (String) — Primary SIP URL.
- `connection_status` (String) — Connection status: `connected`, `disconnected`, `connected_with_issues`, `offline_expired`, `activating`, `pending`, `unknown`, `offline_deep_sleep`.
- `upgrade_channel` (String) — Upgrade channel.
- `location_id` (String) — Location ID.
- `managed_by` (String) — Entity managing the device: `CISCO`, `CUSTOMER`, `PARTNER`.
- `device_platform` (String) — Device platform: `cisco`, `microsoftTeamsRoom`.
- `created` (String) — Creation timestamp (ISO 8601).
- `first_seen` (String) — First seen timestamp (ISO 8601).
- `last_seen` (String) — Last seen timestamp (ISO 8601).

## Import

Devices can be imported using their ID:

```shell
terraform import webex_device.example Y2lzY29zcGFyazovL3...
```
