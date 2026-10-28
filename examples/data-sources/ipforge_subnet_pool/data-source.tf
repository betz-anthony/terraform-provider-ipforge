# Check how much room a pool has left before allocating from it.
data "ipforge_subnet_pool" "linux" {
  subnet_id = 3
  name      = "linux"
}

output "linux_pool_free" {
  value = data.ipforge_subnet_pool.linux.free
}
