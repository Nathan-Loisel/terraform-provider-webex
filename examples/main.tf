terraform {
  required_providers {
    webex = {
      source = "registry.terraform.io/Nathan-Loisel/webex"
    }
  }
}

provider "webex" {
  # token = "your-webex-access-token"  # Or set WEBEX_TOKEN env var
}

# --- Workspace ---

resource "webex_workspace" "boardroom_5f" {
  display_name     = "5th Floor Boardroom"
  type             = "meetingRoom"
  capacity         = 12
  hotdesking_status = "off"
  notes            = "Main boardroom with dual screens"

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

# --- Device (by MAC address) ---

resource "webex_device" "boardroom_codec" {
  mac          = "AA:BB:CC:DD:EE:FF"
  model        = "Cisco Room Kit"
  workspace_id = webex_workspace.boardroom_5f.id
  tags         = ["boardroom", "5th-floor", "managed"]
}

# --- Device Configuration ---

resource "webex_device_configuration" "boardroom_settings" {
  device_id = webex_device.boardroom_codec.id

  configurations = {
    "Standby.Delay"                          = "10"
    "Proximity.Mode"                         = "On"
    "RoomAnalytics.PeopleCountOutOfCall"     = "On"
    "Audio.Ultrasound.MaxVolume"             = "70"
    "UserInterface.OSD.Mode"                 = "Auto"
    "NetworkServices.HTTP.Mode"              = "HTTPS"
    "Video.DefaultMainSource"                = "1"
    "Conference.AutoAnswer.Mode"             = "Off"
  }
}

# --- Data Source: List devices in a workspace ---

data "webex_devices" "boardroom_devices" {
  workspace_id = webex_workspace.boardroom_5f.id
}

output "boardroom_device_count" {
  value = length(data.webex_devices.boardroom_devices.devices)
}

output "workspace_id" {
  value = webex_workspace.boardroom_5f.id
}
