package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"gitlab.com/rackn/terraform-provider-drpv4/drpv4"
)

// Set by goreleaser at link time.
var (
	version = "dev"
	commit  = ""
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with debug logging")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Debug: debug,
	}

	if debug {
		log.Printf("terraform-provider-drp version=%s commit=%s", version, commit)
	}

	err := providerserver.Serve(context.Background(), drpv4.NewProvider(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
