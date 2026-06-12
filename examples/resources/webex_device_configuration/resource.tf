resource "webex_device_configuration" "codec_settings" {
  device_id = webex_device.codec.id

  configurations = {
    "Standby.Delay"                       = "10"
    "Standby.WakeupOnMotionDetection"     = "On"
    "Proximity.Mode"                      = "On"
    "RoomAnalytics.PeopleCountOutOfCall"  = "On"
    "Audio.Ultrasound.MaxVolume"          = "70"
    "Time.Zone"                           = "Europe/Oslo"
  }
}
