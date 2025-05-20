package drpv4

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var testPoolRandomName = fmt.Sprintf("test-%s", randomString(10))

func TestAccResourcePool(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "drp_stage" "test" {
						name = "%s"
						template {
							name = "test"
							contents = <<-EOF
							#!/bin/bash

							echo "test"
							EOF
							path = "/tmp/test"
						}
					}

					resource "drp_workflow" "test" {
						name = "%s"
						description = "test"
						stages = [drp_stage.test.name]
					}

					resource "drp_pool" "test" {
						pool_id = "%s"
						description = "test pool"
						documentation = "test pool"

						allocate_actions {
							workflow = drp_workflow.test.name
						}

						release_actions {
							workflow = drp_workflow.test.name
						}

						enter_actions {
							workflow = drp_workflow.test.name
						}

						exit_actions {
							workflow = drp_workflow.test.name
						}

						autofill {
							max_free = 1
							
							create_parameters = {
								test = jsonencode({
									type = "string"
								})
							}
						}
					}
				`, testPoolRandomName, testPoolRandomName, testPoolRandomName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_pool.test", "pool_id", testPoolRandomName),
					resource.TestCheckResourceAttr("drp_pool.test", "description", "test pool"),
					resource.TestCheckResourceAttr("drp_pool.test", "documentation", "test pool"),
					resource.TestCheckResourceAttr("drp_pool.test", "allocate_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "allocate_actions.0.workflow", testPoolRandomName),
					resource.TestCheckResourceAttr("drp_pool.test", "release_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "release_actions.0.workflow", testPoolRandomName),
					resource.TestCheckResourceAttr("drp_pool.test", "enter_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "enter_actions.0.workflow", testPoolRandomName),
					resource.TestCheckResourceAttr("drp_pool.test", "exit_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "exit_actions.0.workflow", testPoolRandomName),
					resource.TestCheckResourceAttr("drp_pool.test", "autofill.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "autofill.0.max_free", "1"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: fmt.Sprintf(`
					resource "drp_stage" "test" {
						name = "%s"
						template {
							name = "test"
							contents = <<-EOF
							#!/bin/bash

							echo "test"
							EOF
							path = "/tmp/test"
						}
					}

					resource "drp_workflow" "test" {
						name = "%s"
						description = "test"
						stages = [drp_stage.test.name]
					}

					resource "drp_pool" "test" {
						pool_id = "%s"
						description = "test pool"
						documentation = "test pool"

						allocate_actions {
							workflow = drp_workflow.test.name
							remove_parameters = ["test"]
						}

						release_actions {
							workflow = drp_workflow.test.name
						}

						enter_actions {
							workflow = drp_workflow.test.name
						}

						exit_actions {
							workflow = drp_workflow.test.name
						}

						autofill {
							max_free = 1
							
							create_parameters = {
								test = jsonencode({
									type = "string"
								})
							}
						}
					}
				`, testPoolRandomName, testPoolRandomName, testPoolRandomName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_pool.test", "pool_id", testPoolRandomName),
					resource.TestCheckResourceAttr("drp_pool.test", "description", "test pool"),
					resource.TestCheckResourceAttr("drp_pool.test", "documentation", "test pool"),
					resource.TestCheckResourceAttr("drp_pool.test", "allocate_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "allocate_actions.0.workflow", testPoolRandomName),
					resource.TestCheckResourceAttr("drp_pool.test", "release_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "release_actions.0.workflow", testPoolRandomName),
					resource.TestCheckResourceAttr("drp_pool.test", "enter_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "enter_actions.0.workflow", testPoolRandomName),
					resource.TestCheckResourceAttr("drp_pool.test", "exit_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "exit_actions.0.workflow", testPoolRandomName),
					resource.TestCheckResourceAttr("drp_pool.test", "autofill.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "autofill.0.max_free", "1"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: fmt.Sprintf(`
					resource "drp_stage" "test" {
						name = "%s"
						template {
							name = "test"
							contents = <<-EOF
							#!/bin/bash

							echo "test"
							EOF
							path = "/tmp/test"
						}
					}

					resource "drp_workflow" "test" {
						name = "%s"
						description = "test"
						stages = [drp_stage.test.name]
					}

					resource "drp_pool" "test-param" {
						pool_id = "%s-param"
						description = "test pool"
						documentation = "test pool"

						allocate_actions {
							workflow = drp_workflow.test.name
							remove_parameters = ["test"]
							add_parameters = {
								"universal/application" = "image-deploy"
							}
						}

						release_actions {
							workflow = drp_workflow.test.name
							add_parameters = {
								"universal/application" = "hw-only"
							}
						}
					}
				`, testPoolRandomName, testPoolRandomName, testPoolRandomName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_pool.test-param", "allocate_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test-param", "allocate_actions.0.workflow", testPoolRandomName),
					resource.TestCheckResourceAttr("drp_pool.test-param", "allocate_actions.0.add_parameters.%", "1"),
					resource.TestCheckResourceAttr("drp_pool.test-param", `allocate_actions.0.add_parameters.universal/application`, "image-deploy"),
					resource.TestCheckResourceAttr("drp_pool.test-param", "release_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test-param", "release_actions.0.add_parameters.%", "1"),
					resource.TestCheckResourceAttr("drp_pool.test-param", `release_actions.0.add_parameters.universal/application`, "hw-only"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: fmt.Sprintf(`
					resource "drp_param" "test-param" {
						name   = "%s-test"
  						secure = false
  						schema = {
    						type = "boolean"
  						}
					}

					resource "drp_pool" "test-param" {
						pool_id = "%s-param"
						description = "test pool"
						documentation = "test pool"

						allocate_actions {
							workflow = "universal-hardware"
							add_parameters = {
								"universal/application" = "image-deploy"
								"%s-test" = true
							}
						}

						release_actions {
							workflow = "universal-discover"
							add_parameters = {
								"universal/application" = "discover"
								"%s-test" = false
							}
						}
					}
				`, testPoolRandomName, testPoolRandomName, testPoolRandomName, testPoolRandomName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_pool.test-param", "allocate_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test-param", "allocate_actions.0.add_parameters.%", "2"),
					resource.TestCheckResourceAttr("drp_pool.test-param", fmt.Sprintf(`allocate_actions.0.add_parameters.%s-test`, testPoolRandomName), "true"),
					resource.TestCheckResourceAttr("drp_pool.test-param", "release_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test-param", "release_actions.0.add_parameters.%", "2"),
					resource.TestCheckResourceAttr("drp_pool.test-param", fmt.Sprintf(`release_actions.0.add_parameters.%s-test`, testPoolRandomName), "false"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
