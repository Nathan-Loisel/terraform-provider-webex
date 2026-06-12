data "webex_workspaces" "all" {}

output "workspace_names" {
  value = [for ws in data.webex_workspaces.all.workspaces : ws.display_name]
}
