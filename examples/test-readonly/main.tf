terraform {
  required_providers {
    webex = {
      source = "registry.terraform.io/Nathan-Loisel/webex"
    }
  }
}

provider "webex" {}

# Start with a read-only test — list all devices in your org
data "webex_devices" "all" {}

output "total_devices" {
  value = length(data.webex_devices.all.devices)
}

output "device_names" {
  value = [for d in data.webex_devices.all.devices : d.display_name]
}
