package drpv4

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceSubnet(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ResourceName: "drp_subnet.test",
				Config: `
					resource "drp_subnet" "test" {
						name = "test"
						description = "test subnet"
						subnet = "192.168.0.0/24"
						active_start = "192.168.0.1"
						active_end = "192.168.0.255"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_subnet.test", "name", "test"),
					resource.TestCheckResourceAttr("drp_subnet.test", "description", "test subnet"),
					resource.TestCheckResourceAttr("drp_subnet.test", "subnet", "192.168.0.0/24"),
					resource.TestCheckResourceAttr("drp_subnet.test", "active_start", "192.168.0.1"),
					resource.TestCheckResourceAttr("drp_subnet.test", "active_end", "192.168.0.255"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName: "drp_subnet.test",
				Config: `
					resource "drp_subnet" "test" {
						name = "test"
						description = "test subnet"
						subnet = "192.168.0.0/24"
						active_start = "192.168.0.1"
						active_end = "192.168.0.255"
						next_server = "192.168.1.1"

						options {
							code = 1
							value = "255.255.255.0"
						}

						options {
							code = 28
							value = "192.168.0.255"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_subnet.test", "name", "test"),
					resource.TestCheckResourceAttr("drp_subnet.test", "description", "test subnet"),
					resource.TestCheckResourceAttr("drp_subnet.test", "subnet", "192.168.0.0/24"),
					resource.TestCheckResourceAttr("drp_subnet.test", "active_start", "192.168.0.1"),
					resource.TestCheckResourceAttr("drp_subnet.test", "active_end", "192.168.0.255"),
					resource.TestCheckResourceAttr("drp_subnet.test", "next_server", "192.168.1.1"),
					resource.TestCheckResourceAttr("drp_subnet.test", "options.#", "2"),
					resource.TestCheckResourceAttr("drp_subnet.test", "options.0.code", "1"),
					resource.TestCheckResourceAttr("drp_subnet.test", "options.0.value", "255.255.255.0"),
					resource.TestCheckResourceAttr("drp_subnet.test", "options.1.code", "28"),
					resource.TestCheckResourceAttr("drp_subnet.test", "options.1.value", "192.168.0.255"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName: "drp_subnet.test",
				Config: `
					resource "drp_subnet" "test" {
						name = "test1"
						description = "test subnet"
						subnet = "192.168.0.0/24"
						active_start = "192.168.0.1"
						active_end = "192.168.0.255"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_subnet.test", "name", "test1"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName: "drp_subnet.test",
				Config: `
					resource "drp_subnet" "test" {
						name = "test1"
						description = "test subnet"
						subnet = "192.168.0.0/24"
						active_start = "192.168.0.1"
						active_end = "192.168.0.255"
					}
				`,
				ExpectError: regexp.MustCompile("Not Found"),
				// PreConfig: func() {
				// 	t.Log("Removing subnet using DRP API to test terraform error handling")
				// 	client := testAccProvider.Meta().(*Config)
				// 	r, err := client.session.ListModel("subnets")
				// 	if err != nil {
				// 		t.Fatal("failed to list subnets")
				// 	}
				// 	t.Logf("got response: %v", r)

				// 	_, err = client.session.DeleteModel("subnets", "test1")
				// 	if err != nil {
				// 		t.Fatalf("failed to remove subnet using DRP API: %v", err)
				// 	}
				// },
			},
			{
				ResourceName: "drp_subnet.test",
				Config: `
					resource "drp_subnet" "test" {
						name = "test#"
						description = "test subnet"
						subnet = "192.168.0.0/24"
						active_start = "192.168.0.1"
						active_end = "192.168.0.255"
					}
				`,
				ExpectError: regexp.MustCompile("Invalid Name `test#`"),
			},
		},
	})
}
