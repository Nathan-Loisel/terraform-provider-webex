---
page_title: "webex_workspace Resource"
subcategory: ""
description: |-
  Manages a Webex Workspace — a named room with calling, calendar, and device configuration.
---

# webex_workspace

Manages a Webex Workspace. Workspaces represent physical meeting rooms, desks, or collaboration spaces that can have devices, calendars, and calling configurations assigned to them.

Deleting a workspace also deletes all devices associated with it.

## Example Usage

### Basic Meeting Room

```hcl
resource "webex_workspace" "basic" {
  display_name = "Lobby Conference Room"
  type         = "meetingRoom"
  capacity     = 6
}
```

### Full-Featured Boardroom

```hcl
resource "webex_workspace" "boardroom" {
  display_name      = "5th Floor Boardroom"
  type              = "meetingRoom"
  capacity          = 12
  hotdesking_status = "off"
  supported_devices = "collaborationDevices"
  notes             = "Main boardroom — dual 75\" displays"

  calling {
    type = "freeCalling"
  }

  calendar {
    type          = "microsoft"
    email_address = "boardroom-5f@company.com"
  }

  device_hosted_meetings {
    enabled  = true
    site_url = "company.webex.com"
  }
}
```

### Hot Desk

```hcl
resource "webex_workspace" "hotdesk" {
  display_name      = "Open Plan Desk 42"
  type              = "desk"
  capacity          = 1
  hotdesking_status = "on"
}
```

## Schema

### Required

- `display_name` (String) — A friendly name for the workspace.

### Optional

- `org_id` (String) — Organization ID. Only admin users of another organization (such as partners) may use this.
- `location_id` (String) — Location associated with the workspace. **Cannot be changed once configured.** Changing this forces a new resource.
- `floor_id` (String) — Floor associated with the workspace.
- `capacity` (Number) — How many people the workspace is suitable for. Must be 0 or higher.
- `type` (String) — Workspace type. Valid values: `notSet`, `focus`, `huddle`, `meetingRoom`, `open`, `desk`, `other`.
- `sip_address` (String) — SIP address. Only applicable when calling type is `thirdPartySipCalling`.
- `notes` (String) — Notes associated with the workspace.
- `hotdesking_status` (String) — Hot desking status: `on`, `off`.
- `supported_devices` (String) — Supported device type: `collaborationDevices`, `phones`. **Cannot be changed once configured.** Changing this forces a new resource.
- `indoor_navigation_url` (String) — URL of a map locating the workspace.

### Blocks

#### `calling` (Optional)

- `type` (String) — Calling type: `freeCalling`, `hybridCalling`, `webexCalling`, `webexEdgeForDevices`, `thirdPartySipCalling`, `none`.
- `phone_number` (String) — Phone number (webexCalling only).
- `extension` (String) — Extension (webexCalling only).
- `location_id` (String) — Calling location ID (webexCalling only).
- `licenses` (List of String, Read-Only) — Webex Calling licenses assigned to this workspace.

#### `calendar` (Optional)

- `type` (String) — Calendar type: `none`, `google`, `microsoft`.
- `email_address` (String) — Workspace email address.
- `resource_group_id` (String) — Calendar resource group ID.

#### `device_hosted_meetings` (Optional)

- `enabled` (Boolean) — Enable device hosted meetings.
- `site_url` (String) — Webex site URL for device hosted meetings.

### Read-Only

- `id` (String) — Unique identifier for the workspace.
- `created` (String) — Creation timestamp (ISO 8601).
- `device_platform` (String) — Device platform: `cisco`, `microsoftTeamsRoom`.

## Import

Workspaces can be imported using their ID:

```shell
terraform import webex_workspace.example Y2lzY29zcGFyazovL3...
```
