# Claim an address, then publish a DNS A record pointing at it.
# Every dns_record attribute forces replacement (no in-place update).
data "ipforge_subnet" "app" {
  cidr = "10.20.0.0/24"
}

resource "ipforge_allocation" "web01" {
  subnet_id = data.ipforge_subnet.app.id
  hostname  = "web01"
}

resource "ipforge_dns_record" "web01" {
  zone        = "lab.example.com"
  name        = "web01.lab.example.com"
  record_type = "A"
  value       = ipforge_allocation.web01.address
  ttl         = 3600
}
