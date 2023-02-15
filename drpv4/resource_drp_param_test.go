package drpv4

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccParamResourceConfig(name string, description string, documentation string, secure bool) string {
	return fmt.Sprintf(`
	resource "drp_param" "test" {
		name = "%s"
		description = "%s"
		documentation = "%s"
		secure = %t
	}
	`, name, description, documentation, secure)
}

func TestAccParamResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccParamResourceConfig("test", "test", "test", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_param.test", "name", "test"),
					resource.TestCheckResourceAttr("drp_param.test", "description", "test"),
					resource.TestCheckResourceAttr("drp_param.test", "documentation", "test"),
					resource.TestCheckResourceAttr("drp_param.test", "secure", "false"),
				),
			},
		},
	})
}
