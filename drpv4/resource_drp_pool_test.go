package drpv4

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourcePool(t *testing.T) {
	name := fmt.Sprintf("tfpool_%s", accRandomSuffix(10))
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "drp_stage" "test" {
						name = "%[1]s"
						template = [{
							name = "test"
							contents = <<-EOF
							#!/bin/bash

							echo "test"
							EOF
							path = "/tmp/test"
						}]
					}

					resource "drp_workflow" "test" {
						name = "%[1]s"
						description = "test"
						stages = [drp_stage.test.name]
					}

					resource "drp_pool" "test" {
						pool_id = "%[1]s"
						description = "test pool"
						documentation = "test pool"

						allocate_actions = [{
							workflow = drp_workflow.test.name
						}]

						release_actions = [{
							workflow = drp_workflow.test.name
						}]

						enter_actions = [{
							workflow = drp_workflow.test.name
						}]

						exit_actions = [{
							workflow = drp_workflow.test.name
						}]

						autofill = [{
							max_free = 1

							create_parameters = {
								test = jsonencode({
									type = "string"
								})
							}
						}]
					}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_pool.test", "pool_id", name),
					resource.TestCheckResourceAttr("drp_pool.test", "description", "test pool"),
					resource.TestCheckResourceAttr("drp_pool.test", "documentation", "test pool"),
					resource.TestCheckResourceAttr("drp_pool.test", "allocate_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "allocate_actions.0.workflow", name),
					resource.TestCheckResourceAttr("drp_pool.test", "release_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "release_actions.0.workflow", name),
					resource.TestCheckResourceAttr("drp_pool.test", "enter_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "enter_actions.0.workflow", name),
					resource.TestCheckResourceAttr("drp_pool.test", "exit_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "exit_actions.0.workflow", name),
					resource.TestCheckResourceAttr("drp_pool.test", "autofill.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "autofill.0.max_free", "1"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: fmt.Sprintf(`
					resource "drp_stage" "test" {
						name = "%[1]s"
						template = [{
							name = "test"
							contents = <<-EOF
							#!/bin/bash

							echo "test"
							EOF
							path = "/tmp/test"
						}]
					}

					resource "drp_workflow" "test" {
						name = "%[1]s"
						description = "test"
						stages = [drp_stage.test.name]
					}

					resource "drp_pool" "test" {
						pool_id = "%[1]s"
						description = "test pool"
						documentation = "test pool"

						allocate_actions = [{
							workflow = drp_workflow.test.name
							remove_parameters = ["test"]
						}]

						release_actions = [{
							workflow = drp_workflow.test.name
						}]

						enter_actions = [{
							workflow = drp_workflow.test.name
						}]

						exit_actions = [{
							workflow = drp_workflow.test.name
						}]

						autofill = [{
							max_free = 1

							create_parameters = {
								test = jsonencode({
									type = "string"
								})
							}
						}]
					}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_pool.test", "pool_id", name),
					resource.TestCheckResourceAttr("drp_pool.test", "description", "test pool"),
					resource.TestCheckResourceAttr("drp_pool.test", "documentation", "test pool"),
					resource.TestCheckResourceAttr("drp_pool.test", "allocate_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "allocate_actions.0.workflow", name),
					resource.TestCheckResourceAttr("drp_pool.test", "release_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "release_actions.0.workflow", name),
					resource.TestCheckResourceAttr("drp_pool.test", "enter_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "enter_actions.0.workflow", name),
					resource.TestCheckResourceAttr("drp_pool.test", "exit_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "exit_actions.0.workflow", name),
					resource.TestCheckResourceAttr("drp_pool.test", "autofill.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test", "autofill.0.max_free", "1"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: fmt.Sprintf(`
					resource "drp_stage" "test" {
						name = "%[1]s"
						template = [{
							name = "test"
							contents = <<-EOF
							#!/bin/bash

							echo "test"
							EOF
							path = "/tmp/test"
						}]
					}

					resource "drp_workflow" "test" {
						name = "%[1]s"
						description = "test"
						stages = [drp_stage.test.name]
					}

					resource "drp_pool" "test-param" {
						pool_id = "%[1]s-param"
						description = "test pool"
						documentation = "test pool"

						allocate_actions = [{
							workflow = drp_workflow.test.name
							remove_parameters = ["test"]
							add_parameters = {
								"universal/application" = "image-deploy"
							}
						}]

						release_actions = [{
							workflow = drp_workflow.test.name
							add_parameters = {
								"universal/application" = "hw-only"
							}
						}]
					}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_pool.test-param", "allocate_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test-param", "allocate_actions.0.workflow", name),
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
						name   = "%[1]s-test"
						secure = false
						schema = {
							type = "boolean"
						}
					}

					resource "drp_pool" "test-param" {
						pool_id = "%[1]s-param"
						description = "test pool"
						documentation = "test pool"

						allocate_actions = [{
							workflow = "universal-hardware"
							add_parameters = {
								"universal/application" = "image-deploy"
								"%[1]s-test" = true
							}
						}]

						release_actions = [{
							workflow = "universal-discover"
							add_parameters = {
								"universal/application" = "discover"
								"%[1]s-test" = false
							}
						}]
					}
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_pool.test-param", "allocate_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test-param", "allocate_actions.0.add_parameters.%", "2"),
					resource.TestCheckResourceAttr("drp_pool.test-param", fmt.Sprintf(`allocate_actions.0.add_parameters.%s-test`, name), "true"),
					resource.TestCheckResourceAttr("drp_pool.test-param", "release_actions.#", "1"),
					resource.TestCheckResourceAttr("drp_pool.test-param", "release_actions.0.add_parameters.%", "2"),
					resource.TestCheckResourceAttr("drp_pool.test-param", fmt.Sprintf(`release_actions.0.add_parameters.%s-test`, name), "false"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
