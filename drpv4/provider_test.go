package drpv4

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"gitlab.com/rackn/provision/v4/test"
)

const (
	providerConfig string = `
	provider "drp" {
		username = "rocketskates"
		password = "r0ck3tsk4t3s"
		endpoint = "https://127.0.0.1:8092
	}
	`
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"drp": providerserver.NewProtocol6WithError(New("test")()),
	}
)

func init() {
	if os.Getenv("SKIP_TEST_SERVER") == "" {
		log.Println("Starting test server")
		err := test.StartServer(os.TempDir(), 8092)
		if err != nil {
			panic(err)
		}

		time.Sleep(5 * time.Second)
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("RS_ENDPOINT"); v == "" {
		t.Fatal("RS_ENDPOINT must be set for acceptance tests")
	}

	if os.Getenv("RS_TOKEN") == "" && os.Getenv("RS_KEY") == "" {
		t.Fatal("RS_TOKEN or RS_KEY must be set for acceptance tests")
	}
}
