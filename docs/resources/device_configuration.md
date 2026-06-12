---
page_title: "webex_device_configuration Resource"
subcategory: ""
description: |-
  Manages device configurations (xAPI settings) for a Webex device.
---

# webex_device_configuration

Manages device configurations (xAPI settings) for a Webex RoomOS device. Configurations control device behavior — standby timers, proximity, audio, video, network, and UI settings.

This resource only manages the keys you explicitly declare. Other configurations on the device are left untouched. When a key is removed from your configuration, it reverts to the device's factory default. When the entire resource is destroyed, all managed keys revert to defaults.

Configurations can be applied even while the device is offline — they take effect when the device reconnects.

## Example Usage

### Basic Room Settings

```hcl
resource "webex_device_configuration" "room_settings" {
  device_id = webex_device.room_codec.id

  configurations = {
    "Standby.Delay"    = "10"
    "Proximity.Mode"   = "On"
  }
}
```

### Comprehensive Configuration

```hcl
resource "webex_device_configuration" "boardroom" {
  device_id = webex_device.boardroom_codec.id

  configurations = {
    # Standby & Power
    "Standby.Delay"                       = "10"
    "Standby.WakeupOnMotionDetection"     = "On"

    # Proximity & Sensors
    "Proximity.Mode"                      = "On"
    "RoomAnalytics.PeopleCountOutOfCall"  = "On"
    "RoomAnalytics.PeoplePresenceDetector" = "On"

    # Audio
    "Audio.Ultrasound.MaxVolume"          = "70"
    "Audio.DefaultVolume"                 = "50"

    # Video
    "Video.DefaultMainSource"             = "1"
    "Video.Monitors"                      = "Dual"

    # Network
    "NetworkServices.HTTP.Mode"           = "HTTPS"

    # UI
    "UserInterface.OSD.Mode"              = "Auto"

    # Conferencing
    "Conference.AutoAnswer.Mode"          = "Off"
    "Conference.MaxReceiveCallRate"       = "6000"
  }
}
```

## Schema

### Required

- `device_id` (String) — The ID of the device to configure. Changing this forces a new resource.
- `configurations` (Map of String) — Map of configuration keys to values. Keys use dot-separated paths matching the xAPI configuration namespace (e.g., `Standby.Delay`, `Audio.Ultrasound.MaxVolume`). All values are strings. Removing a key from the map reverts it to the device default on the next apply.

### Read-Only

- `id` (String) — Resource identifier (same as `device_id`).

## Configuration Keys

Configuration keys follow the xAPI dot-separated naming convention. Common categories include:

| Category | Example Keys |
|----------|-------------|
| Audio | `Audio.DefaultVolume`, `Audio.Ultrasound.MaxVolume` |
| Video | `Video.DefaultMainSource`, `Video.Monitors` |
| Standby | `Standby.Delay`, `Standby.WakeupOnMotionDetection` |
| Proximity | `Proximity.Mode` |
| Conference | `Conference.AutoAnswer.Mode`, `Conference.MaxReceiveCallRate` |
| NetworkServices | `NetworkServices.HTTP.Mode` |
| RoomAnalytics | `RoomAnalytics.PeopleCountOutOfCall`, `RoomAnalytics.PeoplePresenceDetector` |
| UserInterface | `UserInterface.OSD.Mode` |

Available keys vary by device model and software version. Use the [Webex Device Configurations API](https://developer.webex.com/docs/api/v1/device-configurations) to discover supported keys for a specific device.

## Behavior Notes

- Only keys listed in `configurations` are managed. Other device settings are untouched.
- Removing a key from the map sends a `remove` operation, reverting it to the factory default.
- Destroying this resource reverts **all** managed keys to defaults.
- The API accepts changes even when the device is offline — they apply on reconnect.
- Some configurations may be read-only depending on org policies and device management authority.

## Import

Device configurations can be imported using the device ID. On import, all keys that have a `configured` source value are discovered and added to state:

```shell
terraform import webex_device_configuration.example Y2lzY29zcGFyazovL3...
```
