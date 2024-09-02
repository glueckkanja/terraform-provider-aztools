package main

import (
	"context"
	"flag"
	"log"

	"github.com/glueckkanja/terraform-provider-aztools/internal/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "dev"

	// goreleaser can also pass the specific commit if you want
	// commit  string = ""
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	if debugMode {
		err := plugin.Debug(context.Background(), "github.com/glueckkanja/terraform-provider-aztools/internal/provider",
			&plugin.ServeOpts{
				ProviderFunc: provider.AzTools(version),
			})
		if err != nil {
			log.Println(err.Error())
		}
	} else {
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: provider.AzTools(version)})
	}
}
