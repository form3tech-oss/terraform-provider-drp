package drpv4

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceProfileParam(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "drp_profile_param" "test" {
						profile = "global"
						name = "test"
						schema = {
							type = "string"
						}
					}
					`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_profile_param.test", "name", "test"),
					resource.TestCheckResourceAttr("drp_profile_param.test", "profile", "global"),
					resource.TestCheckResourceAttr("drp_profile_param.test", "schema.type", "string"),
				),
			},
			{
				Config: `
					resource "drp_profile_param" "test" {
						profile = "global"
						name = "test"
						schema = {
							type = "bool"
						}
					}
					`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_profile_param.test", "schema.type", "bool"),
				),
			},
		},
	})
}
