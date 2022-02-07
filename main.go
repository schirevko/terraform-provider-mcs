package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/MailRuCloudSolutions/terraform-provider-vkcs/vkcs"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: vkcs.Provider})
}
