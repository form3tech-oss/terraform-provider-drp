package drpv4

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceReservation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: `
					resource "drp_reservation" "test" {
						address = "192.168.0.1"
						description = "test reservation"
						documentation = "test reservation"
						duration = 86400
						token = "ff:70:81:a9:78:4d"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_reservation.test", "address", "192.168.0.1"),
					resource.TestCheckResourceAttr("drp_reservation.test", "description", "test reservation"),
					resource.TestCheckResourceAttr("drp_reservation.test", "documentation", "test reservation"),
					resource.TestCheckResourceAttr("drp_reservation.test", "duration", "86400"),
					resource.TestCheckResourceAttr("drp_reservation.test", "token", "ff:70:81:a9:78:4d"),
				),
			},
			{
				Config: `
					resource "drp_reservation" "test" {
						address = "192.168.0.1"
						description = "test reservation"
						documentation = "test reservation"
						duration = 86400
						token = "ff:70:81:a9:78:4d"
						next_server = "192.168.1.1"

						options {
							code = 1
							value = "255.255.255.0"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_reservation.test", "next_server", "192.168.1.1"),
				),
			},
			{
				Config: `
					resource "drp_reservation" "test" {
						address = "192.168.0.2"
						description = "test reservation"
						documentation = "test reservation"
						duration = 86400
						token = "ff:70:81:a9:78:4d"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("drp_reservation.test", "address", "192.168.0.2"),
				),
			},
			{
				Config: `
					resource "drp_reservation" "test" {
						address = "192.168.0.256"
						description = "test reservation"
						documentation = "test reservation"
						duration = 86400
						token = "ff:70:81:a9:78:4d"
					}
				`,
				ExpectError: regexp.MustCompile("Empty key not allowed"),
			},
		},
	})
}
