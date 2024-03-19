package drpv4

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTaskResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: `
					resource "drp_task" "test" {
						name = "test"
						description = "test"
						required_params = ["test"]
						optional_params = ["test1"]

						templates {
							name = "test"
							contents = <<-EOF
							#!/bin/bash
							echo "test"
							EOF
							path = "/test.sh"
						}

						extra_claims {
							scope = "*"
							action = "*"
							specific = "*"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("drp_task.test", "name", "test"),
					resource.TestCheckResourceAttr("drp_task.test", "description", "test"),
					resource.TestCheckResourceAttr("drp_task.test", "required_params.#", "1"),
					resource.TestCheckResourceAttr("drp_task.test", "optional_params.#", "1"),
				),
			},
			{
				Config: `
					resource "drp_task" "test" {
						name = "test"
						description = ""
						required_params = ["test"]
						optional_params = ["test1"]

						templates {
							name = "test"
							contents = <<-EOF
							#!/bin/bash
							echo "test1"
							EOF
							path = "/test.sh"
						}

						extra_claims {
							scope = "*"
							action = "*"
							specific = "*"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("drp_task.test", "name", "test"),
					resource.TestCheckResourceAttr("drp_task.test", "description", ""),
					resource.TestCheckResourceAttr("drp_task.test", "required_params.#", "1"),
					resource.TestCheckResourceAttr("drp_task.test", "optional_params.#", "1"),
					resource.TestCheckResourceAttr("drp_task.test", "templates.#", "1"),
					resource.TestCheckResourceAttr("drp_task.test", "templates.0.name", "test"),
					resource.TestCheckResourceAttr("drp_task.test", "templates.0.contents", "#!/bin/bash\necho \"test1\"\n"),
				),
				// ExpectNonEmptyPlan: true,
			},
			{
				Config: `
					resource "drp_task" "test" {
						name = "test"
						description = ""
						required_params = ["test","test2"]
						optional_params = ["test1"]

						templates {
							name = "test"
							contents = <<-EOF
							#!/bin/bash
							echo "test"
							EOF
							path = "/test.sh"
						}

						extra_claims {
							scope = "*"
							action = "*"
							specific = "*"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("drp_task.test", "name", "test"),
					resource.TestCheckResourceAttr("drp_task.test", "description", ""),
					resource.TestCheckResourceAttr("drp_task.test", "required_params.#", "2"),
					resource.TestCheckResourceAttr("drp_task.test", "optional_params.#", "1"),
				),
			},
		},
	})
}
