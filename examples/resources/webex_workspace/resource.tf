resource "webex_workspace" "boardroom" {
  display_name      = "5th Floor Boardroom"
  location_id       = webex_location.oslo.id
  type              = "meetingRoom"
  capacity          = 12
  supported_devices = "collaborationDevices"

  calling {
    type = "freeCalling"
  }

  calendar {
    type          = "microsoft"
    email_address = "boardroom@company.com"
  }

  device_hosted_meetings {
    enabled  = true
    site_url = "company.webex.com"
  }
}
