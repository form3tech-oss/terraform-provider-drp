package drpv4

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccParamResourceConfig(name string, description string, documentation string, schema string, secure bool) string {
	return fmt.Sprintf(`
	resource "drp_param" "test" {
		name = "%s"
		description = "%s"
		documentation = "%s"
		schema = %s
		secure = %t
	}
	`, name, description, documentation, schema, secure)
}

func TestAccParamResource(t *testing.T) {
	teardownTestSuite := setupTestSuite(t)
	defer teardownTestSuite(t)

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccParamResourceConfig("test", "test", "test", "{}", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_param.test", "name", "test"),
					resource.TestCheckResourceAttr("drp_param.test", "description", "test"),
					resource.TestCheckResourceAttr("drp_param.test", "documentation", "test"),
					// resource.TestCheckResourceAttr("drp_param.test", "schema", "{}"),
					resource.TestCheckResourceAttr("drp_param.test", "secure", "false"),
					resource.TestCheckResourceAttrSet("drp_param.test", "schema"),
				),
			},
		},
	})
}
