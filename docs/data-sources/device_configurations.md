---
page_title: "webex_device_configurations Data Source"
subcategory: ""
description: |-
  Reads all device configurations (xAPI settings) for a Webex device.
---

# webex_device_configurations (Data Source)

Reads all device configurations (xAPI settings) for a Webex RoomOS device. Returns current values, factory defaults, and the source of each configuration.

Useful for discovering available configuration keys before managing them with the `webex_device_configuration` resource.

## Example Usage

### Read All Configurations

```hcl
data "webex_device_configurations" "codec" {
  device_id = webex_device.room_codec.id
}

output "standby_delay" {
  value = data.webex_device_configurations.codec.configurations["Standby.Delay"]
}
```

### Filter by Key Prefix

```hcl
data "webex_device_configurations" "audio" {
  device_id  = webex_device.room_codec.id
  key_filter = "Audio.*"
}

output "audio_settings" {
  value = data.webex_device_configurations.audio.configurations
}
```

## Schema

### Required

- `device_id` (String) — The device to read configurations from.

### Optional

- `key_filter` (String) — Filter by key prefix. Supports wildcards (e.g., `Audio.*`, `Standby.*`).

### Read-Only

- `configurations` (Map of String) — Map of configuration key to current effective value.
- `defaults` (Map of String) — Map of configuration key to factory default value.
- `sources` (Map of String) — Map of configuration key to value source (`default` or `configured`).
