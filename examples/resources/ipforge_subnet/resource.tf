resource "ipforge_subnet" "lab" {
  cidr        = "10.20.0.0/24"
  name        = "lab-network"
  description = "Lab subnet managed by Terraform"
}

# Child subnet nested under a parent for hierarchical IPAM.
resource "ipforge_subnet" "lab_mgmt" {
  cidr        = "10.20.0.0/26"
  name        = "lab-management"
  parent_id   = ipforge_subnet.lab.id
  description = "Management range within the lab network"
}
