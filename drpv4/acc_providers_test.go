package drpv4

import (
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"gitlab.com/rackn/provision/v4/test"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"drp": providerserver.NewProtocol6WithError(NewProvider("acc")()),
}

func init() {
	if os.Getenv("SKIP_TEST_SERVER") != "" {
		return
	}
	if os.Getenv("TF_ACC") == "" {
		return
	}
	log.Println("starting DRP test server on :8092")
	if err := test.StartServer(os.TempDir(), 8092); err != nil {
		panic(err)
	}
	time.Sleep(5 * time.Second)
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("RS_ENDPOINT") == "" {
		t.Fatal("RS_ENDPOINT must be set for acceptance tests")
	}
	if os.Getenv("RS_TOKEN") == "" && os.Getenv("RS_KEY") == "" {
		t.Fatal("RS_TOKEN or RS_KEY must be set for acceptance tests")
	}
}

func accRandomSuffix(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
