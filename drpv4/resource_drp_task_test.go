package drpv4

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type TaskResource struct {
	ResourceName   string
	Name           string
	Description    string
	RequiredParams []string
	OptionalParams []string
	Templates      []string
	ExtraClaims    []string
	ExtraRoles     []string
	Prerequisites  []string
}

func testAccTaskResourceConfig(task TaskResource) string {

	requiredParams, err := json.Marshal(task.RequiredParams)
	if err != nil {
		panic(err)
	}

	optionalParams, err := json.Marshal(task.OptionalParams)
	if err != nil {
		panic(err)
	}

	extraRoles, err := json.Marshal(task.ExtraRoles)
	if err != nil {
		panic(err)
	}

	prerequisites, err := json.Marshal(task.Prerequisites)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf(`
		resource "drp_task" "%s" {
			name = "%s"
			description = "%s"
			required_params = %s
			optional_params = %s
			
			%s

			%s

			extra_roles = %s
			prerequisites = %s
		}
	`, task.ResourceName, task.Name, task.Description, requiredParams, optionalParams, task.Templates, task.ExtraClaims, extraRoles, prerequisites)
}

func TestAccTaskResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
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
