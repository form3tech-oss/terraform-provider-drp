package drpv4

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var testWorkflowRandomName = fmt.Sprintf("test-%s", randomString(10))

func TestAccWorkflowResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
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
				`, testWorkflowRandomName, testWorkflowRandomName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_workflow.test", "name", testWorkflowRandomName),
					resource.TestCheckResourceAttr("drp_workflow.test", "description", "test"),
					resource.TestCheckResourceAttr("drp_workflow.test", "stages.#", "1"),
					resource.TestCheckResourceAttr("drp_workflow.test", "stages.0", testWorkflowRandomName),
				),
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
						description = "test1"
						stages = [drp_stage.test.name]
					}`, testWorkflowRandomName, testWorkflowRandomName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_workflow.test", "name", testWorkflowRandomName),
					resource.TestCheckResourceAttr("drp_workflow.test", "description", "test1"),
					resource.TestCheckResourceAttr("drp_workflow.test", "stages.#", "1"),
					resource.TestCheckResourceAttr("drp_workflow.test", "stages.0", testWorkflowRandomName),
				),
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
						name = "%s-1"
						description = "test"
						stages = [drp_stage.test.name]
					}`, testWorkflowRandomName, testWorkflowRandomName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_workflow.test", "name", fmt.Sprintf("%s-1", testWorkflowRandomName)),
					resource.TestCheckResourceAttr("drp_workflow.test", "description", "test"),
					resource.TestCheckResourceAttr("drp_workflow.test", "stages.#", "1"),
					resource.TestCheckResourceAttr("drp_workflow.test", "stages.0", testWorkflowRandomName),
				),
			},
		},
	})
}
