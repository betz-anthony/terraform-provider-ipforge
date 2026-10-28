package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const poolSubnetCfg = `
resource "ipforge_subnet" "test" {
  cidr = "10.253.0.0/24"
  name = "tf-acc-pool-alloc"
}

resource "ipforge_subnet_pool" "a" {
  subnet_id = ipforge_subnet.test.id
  name      = "a"
  start_ip  = "10.253.0.10"
  end_ip    = "10.253.0.11"
}

resource "ipforge_subnet_pool" "b" {
  subnet_id = ipforge_subnet.test.id
  name      = "b"
  start_ip  = "10.253.0.20"
  end_ip    = "10.253.0.21"
}
`

// Allocate from a pool, then switch pools: the existing address belongs to
// the old pool, so the allocation must be replaced, not updated in place.
func TestAccAllocationPoolReplacedOnPoolChange(t *testing.T) {
	alloc := func(pool string) string {
		return poolSubnetCfg + `
resource "ipforge_allocation" "host" {
  subnet_id    = ipforge_subnet.test.id
  hostname     = "tf-acc-pool-host"
  register_dns = false
  pool         = "` + pool + `"
  depends_on   = [ipforge_subnet_pool.a, ipforge_subnet_pool.b]
}

data "ipforge_subnet_pool" "a" {
  subnet_id  = ipforge_subnet.test.id
  name       = "a"
  depends_on = [ipforge_allocation.host]
}`
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: alloc("a"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ipforge_allocation.host", "address", "10.253.0.10"),
					resource.TestCheckResourceAttr("ipforge_allocation.host", "pool", "a"),
					resource.TestCheckResourceAttr("data.ipforge_subnet_pool.a", "total", "2"),
					resource.TestCheckResourceAttr("data.ipforge_subnet_pool.a", "used", "1"),
					resource.TestCheckResourceAttr("data.ipforge_subnet_pool.a", "free", "1"),
				),
			},
			{
				Config: alloc("b"),
				Check:  resource.TestCheckResourceAttr("ipforge_allocation.host", "address", "10.253.0.20"),
			},
		},
	})
}

func TestAccAllocationPoolExhausted(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
resource "ipforge_subnet" "test" {
  cidr = "10.254.0.0/24"
  name = "tf-acc-pool-full"
}

resource "ipforge_subnet_pool" "one" {
  subnet_id = ipforge_subnet.test.id
  name      = "one"
  start_ip  = "10.254.0.10"
  end_ip    = "10.254.0.10"
}

resource "ipforge_allocation" "first" {
  subnet_id = ipforge_subnet.test.id
  hostname  = "tf-acc-first"
  pool      = ipforge_subnet_pool.one.name
}

resource "ipforge_allocation" "second" {
  subnet_id = ipforge_subnet.test.id
  hostname  = "tf-acc-second"
  pool      = ipforge_subnet_pool.one.name
  depends_on = [ipforge_allocation.first]
}`,
			ExpectError: regexp.MustCompile(`exhausted`),
		}},
	})
}
