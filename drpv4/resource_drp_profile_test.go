package drpv4

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceProfile(t *testing.T) {
	profileName := fmt.Sprintf("tfprof_%s", accRandomSuffix(10))
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "drp_profile" "%s" {
						name = "%s"
						description = "My new profile"
					}
				`, profileName, profileName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("drp_profile.%s", profileName), "name", profileName),
					resource.TestCheckResourceAttr(fmt.Sprintf("drp_profile.%s", profileName), "description", "My new profile"),
				),
			},
		},
	})
}
