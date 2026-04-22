package drpv4

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWorkflowResource(t *testing.T) {
	name := fmt.Sprintf("tfstage_%s", accRandomSuffix(10))
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
				`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_workflow.test", "name", name),
					resource.TestCheckResourceAttr("drp_workflow.test", "description", "test"),
					resource.TestCheckResourceAttr("drp_workflow.test", "stages.#", "1"),
					resource.TestCheckResourceAttr("drp_workflow.test", "stages.0", name),
				),
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
						description = "test1"
						stages = [drp_stage.test.name]
					}`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_workflow.test", "name", name),
					resource.TestCheckResourceAttr("drp_workflow.test", "description", "test1"),
					resource.TestCheckResourceAttr("drp_workflow.test", "stages.#", "1"),
					resource.TestCheckResourceAttr("drp_workflow.test", "stages.0", name),
				),
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
						name = "%[1]s-1"
						description = "test"
						stages = [drp_stage.test.name]
					}`, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_workflow.test", "name", fmt.Sprintf("%s-1", name)),
					resource.TestCheckResourceAttr("drp_workflow.test", "description", "test"),
					resource.TestCheckResourceAttr("drp_workflow.test", "stages.#", "1"),
					resource.TestCheckResourceAttr("drp_workflow.test", "stages.0", name),
				),
			},
		},
	})
}
