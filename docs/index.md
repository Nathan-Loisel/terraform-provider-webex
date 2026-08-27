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

The provider supports two authentication methods. Use whichever fits your workflow.

### Method 1 — Static Access Token

Set a pre-obtained access token directly. This is the simplest method for quick usage or testing.

```hcl
provider "webex" {
  token = var.webex_token
}
```

Or via environment variable:

```bash
export WEBEX_TOKEN="your-access-token"
```

**Limitations:**

| Token source | Lifetime | Notes |
|---|---|---|
| Developer portal (temporary token) | **12 hours** | Found at developer.webex.com under "Getting Started". For testing only. |
| Service App access token | **14 days** | Generated from the Org Authorizations section of your Service App page. |

When the token expires, all Terraform operations will fail with a `401 Unauthorized` error. You must manually regenerate the token.

### Method 2 — OAuth Refresh (Recommended for Service Apps)

Set your Service App's `client_id`, `client_secret`, and `refresh_token`. The provider will automatically exchange them for a fresh access token on each Terraform run.

```hcl
provider "webex" {
  client_id     = var.webex_client_id
  client_secret = var.webex_client_secret
  refresh_token = var.webex_refresh_token
}
```

Or via environment variables:

```bash
export WEBEX_CLIENT_ID="your-client-id"
export WEBEX_CLIENT_SECRET="your-client-secret"
export WEBEX_REFRESH_TOKEN="your-refresh-token"
```

**Limitations:**

| Credential | Lifetime | Notes |
|---|---|---|
| Client ID | Permanent | Does not expire. Found on your Service App page. |
| Client Secret | Permanent | Does not expire. Found on your Service App page. |
| Refresh Token | **90 days** | Must be regenerated from the Webex Developer Portal when it expires. |

When the refresh token expires, Terraform will fail with a clear error message explaining how to regenerate it.

### How to Get Your Credentials (Service App Setup)

1. Go to [developer.webex.com](https://developer.webex.com) and sign in.
2. Navigate to **My Webex Apps** → **Create a New App** → **Create a Service App**.
3. Fill in the app name, description, and select the required scopes (see below).
4. After creation, copy the **Client ID** and **Client Secret** from the app page.
5. Click **Request Admin Authorization** — this makes the app visible in your org's Control Hub.
6. A **Full Admin** must approve the Service App in Control Hub.
7. Once authorized, go back to your Service App page on developer.webex.com.
8. Under **Org Authorizations**, select your org from the dropdown.
9. Enter your Client Secret and click **Generate Tokens**.
10. Copy the **Refresh Token** (and optionally the Access Token if using Method 1).

### Which Method Should I Use?

| Scenario | Recommended Method |
|---|---|
| Quick testing / development | Method 1 with a developer portal token (12h) |
| Short-term automation | Method 1 with a Service App access token (14 days) |
| Long-term production use | Method 2 with OAuth refresh (renew every 90 days) |

### Token Expiry Error Messages

When authentication fails, the provider includes guidance in the error message:

- **During OAuth exchange** — If the refresh token has expired, the error will explain that the refresh token (90-day lifetime) needs to be regenerated and how to do it.
- **During API calls** — If the access token is invalid or expired, the error will list all token lifetimes (12h, 14d, 90d) and suggest regeneration steps for both auth methods.

## Rate Limiting

Webex applies its API quota per organisation, so every resource in an apply
competes for the same budget. A plan that touches several workspaces or device
configurations will routinely see `HTTP 429 Too Many Requests`.

The provider retries these itself. A request that fails with `429`, `502`,
`503`, or `504` is retried up to 5 times with exponential backoff and jitter.
When the response carries a `Retry-After` header the provider waits exactly
that long, in either the seconds or HTTP-date form; otherwise the delay
doubles from one second up to a thirty-second ceiling. Backoff is interrupted
if Terraform is cancelled.

Retries are transparent: no configuration is needed, and a request that
eventually succeeds is not reported as an error.

If an apply still fails with `429` after the retries, the organisation is
being throttled harder than the backoff can absorb. Lower Terraform's
concurrency:

```console
$ terraform apply -parallelism=2
```

Only `429`, `502`, `503`, and `504` are retried. Other `4xx` responses are
returned immediately, since retrying a rejected request will not change the
outcome.

## Required Scopes

| Scope | Purpose |
|-------|---------|
| `spark-admin:locations_read` | Read location and floor details |
| `spark-admin:locations_write` | Create, update, delete locations and floors |
| `spark-admin:workspaces_read` | Read workspace details |
| `spark-admin:workspaces_write` | Create, update, delete workspaces |
| `spark-admin:devices_read` | Read device details and configurations |
| `spark-admin:devices_write` | Create, delete devices and modify configurations |

## Example Usage

### With OAuth refresh (recommended)

```hcl
provider "webex" {
  client_id     = var.webex_client_id
  client_secret = var.webex_client_secret
  refresh_token = var.webex_refresh_token
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
```

### With static token

```hcl
provider "webex" {
  token = var.webex_token
}

resource "webex_device_configuration" "room_settings" {
  device_id = "device-id-here"

  configurations = {
    "Standby.Delay" = "10"
    "Proximity.Mode" = "On"
  }
}
```

## Schema

### Optional

- `token` (String, Sensitive) — Webex API access token. Can also be set via the `WEBEX_TOKEN` environment variable. Developer portal tokens expire after 12 hours; Service App access tokens expire after 14 days.
- `client_id` (String) — Webex Service App Client ID. Can also be set via the `WEBEX_CLIENT_ID` environment variable. Required together with `client_secret` and `refresh_token`.
- `client_secret` (String, Sensitive) — Webex Service App Client Secret. Can also be set via the `WEBEX_CLIENT_SECRET` environment variable. Required together with `client_id` and `refresh_token`.
- `refresh_token` (String, Sensitive) — Webex Service App Refresh Token. Can also be set via the `WEBEX_REFRESH_TOKEN` environment variable. Expires after 90 days. Required together with `client_id` and `client_secret`.
- `base_url` (String) — Webex API base URL. Defaults to `https://webexapis.com/v1`.
