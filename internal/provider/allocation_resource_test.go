package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAllocationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "ipforge_subnet" "test" {
  cidr = "10.251.0.0/24"
  name = "tf-acc-alloc"
}

resource "ipforge_allocation" "test" {
  subnet_id    = ipforge_subnet.test.id
  hostname     = "tf-acc-host"
  register_dns = false
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ipforge_allocation.test", "id"),
					resource.TestCheckResourceAttrSet("ipforge_allocation.test", "address"),
				),
			},
		},
	})
}
