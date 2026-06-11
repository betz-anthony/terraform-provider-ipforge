resource "ipforge_subnet" "lab" {
  cidr = "10.20.0.0/24"
  name = "lab-network"
}

resource "ipforge_address" "gateway" {
  address     = "10.20.0.1"
  subnet_id   = ipforge_subnet.lab.id
  hostname    = "lab-gw"
  status      = "reserved"
  mac_address = "00:11:22:33:44:55"
  description = "Lab default gateway"
}
