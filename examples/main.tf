# End-to-end example: discover a subnet, claim an address from it, and publish
# a matching DNS A record. Token is read from the IPFORGE_TOKEN env var.
terraform {
  required_providers {
    ipforge = {
      source = "betz-anthony/ipforge"
    }
  }
}

provider "ipforge" {
  url = "https://ipforge.example.com"
  # token sourced from IPFORGE_TOKEN
}

data "ipforge_subnet" "app" {
  cidr = "10.20.0.0/24"
}

resource "ipforge_allocation" "web01" {
  subnet_id    = data.ipforge_subnet.app.id
  hostname     = "web01"
  description  = "Web frontend, allocated by Terraform"
  register_dns = false
}

resource "ipforge_dns_record" "web01" {
  zone        = "lab.example.com"
  name        = "web01.lab.example.com"
  record_type = "A"
  value       = ipforge_allocation.web01.address
  ttl         = 3600
}

output "web01_ip" {
  value = ipforge_allocation.web01.address
}
