package drpv4

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var profileParamName = fmt.Sprintf("test-%s", randomString(10))

func TestAccResourceProfileParam(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "drp_profile_param" "%s" {
						profile = "global"
						name = "test1"
					}
				`, profileParamName),
				ExpectError: regexp.MustCompile("Invalid combination of arguments"),
			},
			{
				Config: fmt.Sprintf(`
					resource "drp_profile_param" "%s" {
						profile = "global"
						name = "test"
						value = "test"
					}
					`, profileParamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("drp_profile_param.%s", profileParamName), "name", "test"),
					resource.TestCheckResourceAttr(fmt.Sprintf("drp_profile_param.%s", profileParamName), "profile", "global"),
					resource.TestCheckResourceAttr(fmt.Sprintf("drp_profile_param.%s", profileParamName), "value", "test"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "drp_profile_param" "%s" {
						profile = "global"
						name = "test"
						value = "test2"
					}
					`, profileParamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("drp_profile_param.%s", profileParamName), "value", "test2"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "drp_param" "%s" {
						name = "my_password"
						description = "Password"
						schema = {
							type = "string"
						}
						secure = true
					}

					resource "drp_profile_param" "%s" {
						profile = "global"
						name = drp_param.%s.name
						secure_value = "test2"
					}
				`, profileParamName, profileParamName, profileParamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("drp_profile_param.%s", profileParamName), "secure_value", "test2"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "drp_param" "%s_bool" {
						name = "my_boolean"
						description = "Boolean"
						schema = {
							type = "boolean"
						}
					}

					resource "drp_profile_param" "%s_bool" {
						profile = "global"
						name = drp_param.%s_bool.name
						value = "true"
					}
				`, profileParamName, profileParamName, profileParamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("drp_profile_param.%s_bool", profileParamName), "value", "true"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "drp_param" "%s_integer" {
						name = "my_integer"
						description = "Integer"
						schema = {
							type = "integer"
						}
					}

					resource "drp_profile_param" "%s_integer" {
						profile = "global"
						name = drp_param.%s_integer.name
						value = "123456789"
					}
				`, profileParamName, profileParamName, profileParamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("drp_profile_param.%s_integer", profileParamName), "value", "123456789"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "drp_param" "%s_number" {
						name = "my_number"
						description = "Number"
						schema = {
							type = "number"
						}
					}

					resource "drp_profile_param" "%s_number" {
						profile = "global"
						name = drp_param.%s_number.name
						value = "1.234567"
					}
				`, profileParamName, profileParamName, profileParamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("drp_profile_param.%s_number", profileParamName), "value", "1.234567"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "drp_param" "%s" {
						name = "my_password"
						description = "Password"
						schema = {
							type = "string"
						}
						secure = true
					}

					resource "drp_profile_param" "%s" {
						profile = "global"
						name = drp_param.%s.name
						value = "test2"
					}
				`, profileParamName, profileParamName, profileParamName),
				ExpectError: regexp.MustCompile("param my_password is secure, use secure_value instead"),
			},
		},
	})
}
