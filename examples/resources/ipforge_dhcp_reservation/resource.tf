resource "ipforge_dhcp_reservation" "printer" {
  scope_id    = "10.20.0.0"
  ip_address  = "10.20.0.50"
  mac_address = "aa:bb:cc:dd:ee:ff"
  name        = "lab-printer"
  description = "Static DHCP reservation managed by Terraform"
}
