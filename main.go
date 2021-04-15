package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"

	"github.com/mubi/terraform-provider-citrixitm/citrixitm"
)

func main() {
	plugin.Serve(
		&plugin.ServeOpts{
			ProviderFunc: func() terraform.ResourceProvider {
				return citrixitm.Provider()
			},
		})
}
