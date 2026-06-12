data "webex_devices" "boardroom" {
  workspace_id = webex_workspace.boardroom.id
}

output "device_names" {
  value = [for d in data.webex_devices.boardroom.devices : d.display_name]
}
