---
page_title: "Webex Provider"
subcategory: ""
description: |-
  Terraform provider for managing Cisco Webex Workspaces, Devices, and Device Configurations.
---

# Webex Provider

The Webex provider enables Terraform to manage Cisco Webex room infrastructure — workspaces, devices, and device configurations (xAPI settings).

This provider targets the admin-level Webex REST APIs for physical room/device management, not the collaboration APIs (messaging, meetings, etc.).

## Authentication

The provider requires a Webex API access token with admin-level scopes. You can obtain one by creating a [Service App](https://developer.webex.com/docs/service-app) or an [Integration](https://developer.webex.com/docs/integrations) in the Webex Developer Portal.

### Required Scopes

| Scope | Purpose |
|-------|---------|
| `spark-admin:locations_read` | Read location and floor details |
| `spark-admin:locations_write` | Create, update, delete locations and floors |
| `spark-admin:workspaces_read` | Read workspace details |
| `spark-admin:workspaces_write` | Create, update, delete workspaces |
| `spark-admin:devices_read` | Read device details and configurations |
| `spark-admin:devices_write` | Create, delete devices and modify configurations |

### Setting the Token

The token can be set via the provider configuration or the `WEBEX_TOKEN` environment variable:

```bash
export WEBEX_TOKEN="your-webex-access-token"
```

## Example Usage

```hcl
provider "webex" {
  # token = "..."  # Or set WEBEX_TOKEN env var
}

resource "webex_workspace" "meeting_room" {
  display_name = "3rd Floor Meeting Room"
  type         = "meetingRoom"
  capacity     = 8

  calling {
    type = "freeCalling"
  }

  calendar {
    type          = "microsoft"
    email_address = "room-3f@company.com"
  }
}

resource "webex_device_configuration" "room_settings" {
  device_id = "device-id-here"

  configurations = {
    "Standby.Delay"      = "10"
    "Proximity.Mode"     = "On"
  }
}
```

## Schema

### Optional

- `token` (String, Sensitive) — Webex API access token. Can also be set via the `WEBEX_TOKEN` environment variable.
- `base_url` (String) — Webex API base URL. Defaults to `https://webexapis.com/v1`.
