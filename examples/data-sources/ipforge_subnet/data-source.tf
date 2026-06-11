# Look up a subnet by CIDR (or by name) to reference its id elsewhere.
data "ipforge_subnet" "app" {
  cidr = "10.20.0.0/24"
}

output "app_subnet_id" {
  value = data.ipforge_subnet.app.id
}
