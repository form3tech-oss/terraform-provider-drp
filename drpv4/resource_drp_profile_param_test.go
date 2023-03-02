package drpv4

import (
	"regexp"
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
						name = "test1"
					}
				`,
				ExpectError: regexp.MustCompile("Invalid combination of arguments"),
			},
			{
				Config: `
					resource "drp_profile_param" "test" {
						profile = "global"
						name = "test"
						value = "test"
					}
					`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_profile_param.test", "name", "test"),
					resource.TestCheckResourceAttr("drp_profile_param.test", "profile", "global"),
					resource.TestCheckResourceAttr("drp_profile_param.test", "value", "test"),
				),
			},
			{
				Config: `
					resource "drp_profile_param" "test" {
						profile = "global"
						name = "test"
						value = "test2"
					}
					`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_profile_param.test", "value", "test2"),
				),
			},
			{
				Config: `
					resource "drp_param" "test" {
						name = "my_password"
						description = "Password"
						schema = {
							type = "string"
						}
						secure = true
					}

					resource "drp_profile_param" "test" {
						profile = "global"
						name = drp_param.test.name
						secure_value = "test2"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_profile_param.test", "secure_value", "test2"),
				),
			},
			{
				Config: `
					resource "drp_param" "test" {
						name = "my_password"
						description = "Password"
						schema = {
							type = "string"
						}
						secure = true
					}

					resource "drp_profile_param" "test" {
						profile = "global"
						name = drp_param.test.name
						value = "test2"
					}
				`,
				ExpectError: regexp.MustCompile("param my_password is secure, use secure_value instead"),
			},
		},
	})
}
