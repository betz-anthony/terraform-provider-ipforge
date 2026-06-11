# Look up an existing address by IP.
data "ipforge_address" "gateway" {
  ip = "10.20.0.1"
}

output "gateway_hostname" {
  value = data.ipforge_address.gateway.hostname
}
