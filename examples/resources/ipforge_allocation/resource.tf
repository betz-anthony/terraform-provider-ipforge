# Look up an existing subnet, then claim the next free address in it.
# The allocation is idempotent by hostname and is released on destroy.
data "ipforge_subnet" "app" {
  cidr = "10.20.0.0/24"
}

resource "ipforge_allocation" "web01" {
  subnet_id    = data.ipforge_subnet.app.id
  hostname     = "web01"
  description  = "Web frontend, allocated by Terraform"
  register_dns = true
  dns_zone     = "lab.example.com"
}

output "web01_ip" {
  value = ipforge_allocation.web01.address
}
