package drpv4

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type TemplateResource struct {
	ResourceName string
	ID           string
	Description  string
	Contents     string
	StartDelim   string
	EndDelim     string
}

func testAccTemplateResourceConfig(template TemplateResource) string {
	return fmt.Sprintf(`
		resource "drp_template" "%s" {
			template_id = "%s"
			description = "%s"
			contents = "%s"
			start_delimiter = "%s"
			end_delimiter = "%s"
		}
	`, template.ResourceName, template.ID, template.Description, template.Contents, template.StartDelim, template.EndDelim)
}

func TestAccTemplateResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
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
					ID:           "test",
					Description:  "test",
					Contents:     "test",
					StartDelim:   "[[",
					EndDelim:     "]]",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("drp_template.test", "start_delimiter", "[["),
					resource.TestCheckResourceAttr("drp_template.test", "end_delimiter", "]]"),
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
