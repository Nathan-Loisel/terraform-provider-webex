data "webex_workspace" "boardroom" {
  display_name = "5th Floor Boardroom"
}

output "boardroom_id" {
  value = data.webex_workspace.boardroom.id
}
