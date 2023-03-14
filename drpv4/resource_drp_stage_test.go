package drpv4

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccStageResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: `
					resource "drp_stage" "test" {
						name = "test"
						description = "test"
						params = {
							test = "test"
						}
						
						optional_params = ["test"]
						runner_wait = true
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_stage.test", "name", "test"),
					resource.TestCheckResourceAttr("drp_stage.test", "description", "test"),
					resource.TestCheckResourceAttr("drp_stage.test", "params.%", "1"),
					resource.TestCheckResourceAttr("drp_stage.test", "params.test", "test"),
					resource.TestCheckResourceAttr("drp_stage.test", "runner_wait", "true"),
				),
			},
			{
				Config: `
					resource "drp_stage" "test" {
						name = "test"
						description = "test"
						params = {
							test = "test"
						}
						
						optional_params = ["test"]
						runner_wait = true
						
						template {
							name = "test"
							contents = <<-EOF
							#!/bin/bash

							echo "test"
							EOF
							path = "/tmp/test"
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_stage.test", "name", "test"),
					resource.TestCheckResourceAttr("drp_stage.test", "description", "test"),
					resource.TestCheckResourceAttr("drp_stage.test", "params.%", "1"),
					resource.TestCheckResourceAttr("drp_stage.test", "params.test", "test"),
					resource.TestCheckResourceAttr("drp_stage.test", "runner_wait", "true"),
					resource.TestCheckResourceAttr("drp_stage.test", "template.#", "1"),
					resource.TestCheckResourceAttr("drp_stage.test", "template.0.name", "test"),
					resource.TestCheckResourceAttr("drp_stage.test", "template.0.path", "/tmp/test"),
					resource.TestCheckResourceAttr("drp_stage.test", "template.0.contents", "#!/bin/bash\n\necho \"test\"\n"),
				),
			},
			{
				Config: `
					resource "drp_stage" "test" {
						name = "test"
						description = "test"
						params = {
							test = "test"
						}
						
						runner_wait = true
						
						template {
							name = "test1"
							contents = <<-EOF
							#!/bin/bash

							echo "test"
							EOF
							path = "/tmp/test"
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_stage.test", "name", "test"),
					resource.TestCheckResourceAttr("drp_stage.test", "description", "test"),
					resource.TestCheckResourceAttr("drp_stage.test", "params.%", "1"),
					resource.TestCheckResourceAttr("drp_stage.test", "params.test", "test"),
					resource.TestCheckResourceAttr("drp_stage.test", "runner_wait", "true"),
					resource.TestCheckTypeSetElemAttr("drp_stage.test", "optional_params.*", "0"),
				),
			},
			{
				Config: `
					resource "drp_stage" "test" {
						name = "test#"
						description = "test"
						params = {
							test = "test"
						}
						
						optional_params = ["test"]
						runner_wait = true
					}`,
				ExpectError: regexp.MustCompile("Invalid Name `test#`"),
			},
		},
	})
}
