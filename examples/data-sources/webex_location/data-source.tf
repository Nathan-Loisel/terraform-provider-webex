data "webex_location" "oslo" {
  name = "Oslo HQ"
}

output "oslo_timezone" {
  value = data.webex_location.oslo.time_zone
}
