package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"net/http"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"vsts_repository":       resourceRepository(),
			"vsts_project":          resourceProject(),
			"vsts_service_endpoint": resourceServiceEndpoint(),
		},
		ConfigureFunc: providerConfigure,
		Schema: map[string]*schema.Schema{
			"account": {
				Required:    true,
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("VSTS_ACCOUNT", nil),
			},
			"username": {
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("VSTS_USER", nil),
				Required:    true,
			},
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VSTS_TOKEN", nil),
			},
		},
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	client := &VSTSClient{
		Account:    d.Get("account").(string),
		Token:      d.Get("token").(string),
		HTTPClient: &http.Client{},
	}

	return client, nil
}
