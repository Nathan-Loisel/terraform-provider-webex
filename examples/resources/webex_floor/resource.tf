resource "webex_floor" "ground" {
  location_id  = webex_location.oslo.id
  floor_number = 1
  display_name = "Ground Floor"
}
