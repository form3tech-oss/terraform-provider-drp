package drpv4

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/rackn/provision/v4/test"
)

var testAccProvider *schema.Provider
var testAccProviders map[string]*schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"drp": testAccProvider,
	}

	err := test.StartServer(os.TempDir(), 8092)
	if err != nil {
		panic(err)
	}
	defer test.StopServer()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("RS_ENDPOINT"); v == "" {
		t.Fatal("RS_ENDPOINT must be set for acceptance tests")
	}

	if os.Getenv("RS_TOKEN") == "" && os.Getenv("RS_KEY") == "" {
		t.Fatal("RS_TOKEN or RS_KEY must be set for acceptance tests")
	}
}
