package drpv4

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type TemplateResource struct {
	ResourceName string
	ID           string
	Description  string
	Contents     string
}

func testAccTemplateResourceConfig(template TemplateResource) string {
	return fmt.Sprintf(`
		resource "drp_template" "%s" {
			template_id = "%s"
			description = "%s"
			contents = "%s"
		}
	`, template.ResourceName, template.ID, template.Description, template.Contents)
}

func TestAccTemplateResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateResourceConfig(TemplateResource{
					ResourceName: "test",
					ID:           "test",
					Description:  "test",
					Contents:     "test",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("drp_template.test", "id", "test"),
					resource.TestCheckResourceAttr("drp_template.test", "description", "test"),
					resource.TestCheckResourceAttr("drp_template.test", "contents", "test"),
				),
			},
			{
				Config: testAccTemplateResourceConfig(TemplateResource{
					ResourceName: "test",
					ID:           "test#",
					Description:  "test",
					Contents:     "test",
				}),
				ExpectError: regexp.MustCompile("Invalid ID `test#`"),
			},
		},
	})
}
