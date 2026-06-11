resource "ipforge_vlan" "lab" {
  vlan_id     = 120
  name        = "lab"
  description = "Lab VLAN managed by Terraform"
}
