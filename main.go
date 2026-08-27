// terraform-provider-freshdesk is the Terraform provider for Freshdesk.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/slop-place/terraform-provider-freshdesk/internal/provider"
)

// version is set at build time with -ldflags.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false,
		"run the provider with support for debuggers such as delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/slop-place/freshdesk",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
