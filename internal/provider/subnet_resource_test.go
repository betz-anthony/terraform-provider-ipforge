package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccPreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 and IPFORGE_URL/IPFORGE_TOKEN to run acceptance tests")
	}
	for _, k := range []string{"IPFORGE_URL", "IPFORGE_TOKEN"} {
		if os.Getenv(k) == "" {
			t.Fatalf("%s must be set for acceptance tests", k)
		}
	}
}

func TestAccSubnetResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "ipforge_subnet" "test" {
  cidr = "10.250.0.0/24"
  name = "tf-acc-test"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ipforge_subnet.test", "cidr", "10.250.0.0/24"),
					resource.TestCheckResourceAttrSet("ipforge_subnet.test", "id"),
				),
			},
			{ResourceName: "ipforge_subnet.test", ImportState: true, ImportStateVerify: true},
		},
	})
}
