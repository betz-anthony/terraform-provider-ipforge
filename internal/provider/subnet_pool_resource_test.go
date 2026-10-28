package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSubnetPoolResource(t *testing.T) {
	cfg := func(end, extra string) string {
		return fmt.Sprintf(`
resource "ipforge_subnet" "test" {
  cidr = "10.252.0.0/24"
  name = "tf-acc-pool"
}

resource "ipforge_subnet_pool" "linux" {
  subnet_id = ipforge_subnet.test.id
  name      = "linux"
  start_ip  = "10.252.0.100"
  end_ip    = "%s"%s
}`, end, extra)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cfg("10.252.0.109", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ipforge_subnet_pool.linux", "id"),
					resource.TestCheckResourceAttr("ipforge_subnet_pool.linux", "name", "linux"),
					resource.TestCheckResourceAttr("ipforge_subnet_pool.linux", "end_ip", "10.252.0.109"),
				),
			},
			{
				// In-place update: range and description change, id is kept.
				Config: cfg("10.252.0.119", "\n  description = \"linux hosts\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ipforge_subnet_pool.linux", "end_ip", "10.252.0.119"),
					resource.TestCheckResourceAttr("ipforge_subnet_pool.linux", "description", "linux hosts"),
				),
			},
			{
				ResourceName:      "ipforge_subnet_pool.linux",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["ipforge_subnet_pool.linux"]
					return rs.Primary.Attributes["subnet_id"] + "/linux", nil
				},
			},
		},
	})
}

func TestAccSubnetPoolRejectsUppercaseName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
resource "ipforge_subnet_pool" "bad" {
  subnet_id = 1
  name      = "Linux"
  start_ip  = "10.0.0.1"
  end_ip    = "10.0.0.2"
}`,
			ExpectError: regexp.MustCompile(`Invalid pool name`),
		}},
	})
}
