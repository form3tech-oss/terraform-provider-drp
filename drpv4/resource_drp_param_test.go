package drpv4

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type ParamResource struct {
	ResourceName  string
	Name          string
	Description   string
	Documentation string
	Schema        string
	Secure        bool
}

func testAccParamResourceConfig(param ParamResource) string {
	return fmt.Sprintf(`
		resource "drp_param" "%s" {
			name = "%s"
			description = "%s"
			documentation = "%s"
			schema = %s
			secure = %t
		}
	`, param.ResourceName, param.Name, param.Description, param.Documentation, param.Schema, param.Secure)
}

func TestAccParamResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccParamResourceConfig(ParamResource{
					ResourceName:  "test",
					Name:          "test",
					Description:   "test",
					Documentation: "test",
					Schema:        `{"type": "string"}`,
					Secure:        false,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("drp_param.test", "name", "test"),
					resource.TestCheckResourceAttr("drp_param.test", "description", "test"),
					resource.TestCheckResourceAttr("drp_param.test", "documentation", "test"),
					resource.TestCheckResourceAttr("drp_param.test", "schema.%", "1"),
					resource.TestCheckResourceAttr("drp_param.test", "schema.type", "string"),
					resource.TestCheckResourceAttr("drp_param.test", "secure", "false"),
				),
			},
			{
				Config: testAccParamResourceConfig(ParamResource{
					ResourceName:  "test",
					Name:          "test123",
					Documentation: "test",
					Schema:        `{"type": "string"}`,
					Secure:        false,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("drp_param.test", "name", "test123"),
					resource.TestCheckResourceAttr("drp_param.test", "documentation", "test"),
					resource.TestCheckResourceAttr("drp_param.test", "schema.%", "1"),
					resource.TestCheckResourceAttr("drp_param.test", "schema.type", "string"),
					resource.TestCheckResourceAttr("drp_param.test", "secure", "false"),
				),
			},
			{
				Config: testAccParamResourceConfig(ParamResource{
					ResourceName:  "test",
					Name:          "test123",
					Description:   "testing some change",
					Documentation: "test",
					Schema:        `{"type": "string"}`,
					Secure:        false,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("drp_param.test", "name", "test123"),
					resource.TestCheckResourceAttr("drp_param.test", "description", "testing some change"),
					resource.TestCheckResourceAttr("drp_param.test", "documentation", "test"),
					resource.TestCheckResourceAttr("drp_param.test", "schema.%", "1"),
					resource.TestCheckResourceAttr("drp_param.test", "schema.type", "string"),
					resource.TestCheckResourceAttr("drp_param.test", "secure", "false"),
				),
			},
			{
				Config: testAccParamResourceConfig(ParamResource{
					ResourceName: "test2",
					Name:         "test2",
					Description:  "",
					Schema:       `{"type": "string"}`,
					Secure:       true,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_param.test2", "name", "test2"),
					resource.TestCheckResourceAttr("drp_param.test2", "description", ""),
					resource.TestCheckResourceAttr("drp_param.test2", "documentation", ""),
					resource.TestCheckResourceAttr("drp_param.test2", "schema.%", "1"),
					resource.TestCheckResourceAttr("drp_param.test2", "schema.type", "string"),
					resource.TestCheckResourceAttr("drp_param.test2", "secure", "true"),
				),
			},
			{
				Config: testAccParamResourceConfig(ParamResource{
					ResourceName: "test3",
					Name:         "test3",
					Description:  "",
					Schema:       `{"type": "string", "default": "testing"}`,
					Secure:       false,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_param.test3", "name", "test3"),
					resource.TestCheckResourceAttr("drp_param.test3", "description", ""),
					resource.TestCheckResourceAttr("drp_param.test3", "documentation", ""),
					resource.TestCheckResourceAttr("drp_param.test3", "schema.%", "2"),
					resource.TestCheckResourceAttr("drp_param.test3", "schema.type", "string"),
					resource.TestCheckResourceAttr("drp_param.test3", "schema.default", "testing"),
					resource.TestCheckResourceAttr("drp_param.test3", "secure", "false"),
				),
			},
			{
				Config: testAccParamResourceConfig(ParamResource{
					ResourceName: "test4",
					Name:         "test4#",
					Description:  "",
					Schema:       `{}`,
					Secure:       false,
				}),
				ExpectError: regexp.MustCompile(".*Invalid Name `test4#`"),
			},
		},
	})
}
