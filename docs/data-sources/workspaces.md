---
page_title: "webex_workspaces Data Source"
subcategory: ""
description: |-
  Lists Webex Workspaces, optionally filtered by display name.
---

# webex_workspaces (Data Source)

Lists Webex Workspaces in your organization. Can filter by display name (partial match).

## Example Usage

### List All Workspaces

```hcl
data "webex_workspaces" "all" {}

output "workspace_count" {
  value = length(data.webex_workspaces.all.workspaces)
}
```

### Filter by Name

```hcl
data "webex_workspaces" "meeting_rooms" {
  display_name = "Meeting"
}

output "meeting_rooms" {
  value = [for ws in data.webex_workspaces.meeting_rooms.workspaces : ws.display_name]
}
```

## Schema

### Optional

- `display_name` (String) — Filter workspaces by display name (partial match).

### Read-Only

- `workspaces` (List of Object) — List of workspaces. Each workspace contains:
  - `id` (String) — Workspace ID.
  - `display_name` (String) — Display name.
  - `org_id` (String) — Organization ID.
  - `type` (String) — Workspace type.
  - `capacity` (Number) — Room capacity.
  - `location_id` (String) — Location ID.
  - `floor_id` (String) — Floor ID.
  - `sip_address` (String) — SIP address.
  - `created` (String) — Creation timestamp.
  - `notes` (String) — Notes.
  - `hotdesking_status` (String) — Hot desking status.
  - `supported_devices` (String) — Supported device type.
  - `device_platform` (String) — Device platform.
  - `indoor_navigation_url` (String) — Indoor navigation map URL.
  - `calling` — Calling configuration:
    - `type` (String) — Calling type.
    - `phone_number` (String) — Phone number.
    - `extension` (String) — Extension.
    - `location_id` (String) — Calling location ID.
    - `licenses` (List of String) — Webex Calling licenses.
  - `calendar` — Calendar configuration:
    - `type` (String) — Calendar type.
    - `email_address` (String) — Calendar email address.
    - `resource_group_id` (String) — Calendar resource group ID.
  - `device_hosted_meetings` — Device hosted meetings:
    - `enabled` (Boolean) — Whether enabled.
    - `site_url` (String) — Webex site URL.
