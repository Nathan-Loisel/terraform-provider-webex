data "webex_device_configurations" "codec" {
  device_id = webex_device.codec.id
}

output "standby_delay" {
  value = data.webex_device_configurations.codec.configurations["Standby.Delay"]
}
