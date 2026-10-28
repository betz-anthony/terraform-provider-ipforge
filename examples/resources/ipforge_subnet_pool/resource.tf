# Split one subnet into per-workload ranges, then claim the next free
# address from a specific range.
data "ipforge_subnet" "servers" {
  cidr = "10.20.0.0/22"
}

resource "ipforge_subnet_pool" "linux" {
  subnet_id   = data.ipforge_subnet.servers.id
  name        = "linux"
  start_ip    = "10.20.3.1"
  end_ip      = "10.20.3.254"
  description = "Linux hosts"
}

# The allocation takes its address from the pool. Changing `pool` later
# replaces the allocation, because the old address belongs to the old range.
resource "ipforge_allocation" "web01" {
  subnet_id = data.ipforge_subnet.servers.id
  hostname  = "web01"
  pool      = ipforge_subnet_pool.linux.name
}
