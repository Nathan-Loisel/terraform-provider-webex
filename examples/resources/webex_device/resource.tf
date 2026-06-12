resource "webex_device" "codec" {
  mac          = "AA:BB:CC:DD:EE:FF"
  model        = "Cisco Room Kit EQ"
  workspace_id = webex_workspace.boardroom.id
  tags         = ["managed", "5th-floor"]
}
