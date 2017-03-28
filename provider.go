package main

import (
	"net/http"

	"github.com/hashicorp/terraform/helper/schema"
)

// Provider creates the Instaclustr provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("INSTACLUSTR_ACCESS_KEY", ""),
				Description: "Instaclustr user name used for api access",
			},
			"secret_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("INSTACLUSTR_SECRET_KEY", ""),
				Description: "Instaclustr key used for api access",
			},
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INSTACLUSTR_URL", "https://api.instaclustr.com/provisioning/v1"),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"instaclustr_firewall_rule": resourceFirewallRule(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {

	config := Config{
		AccessKey: d.Get("access_key").(string),
		SecretKey: d.Get("secret_key").(string),
		URL:       d.Get("url").(string),
	}

	return &InstaclustrClient{
		config: config,
		client: &http.Client{},
	}, nil
}
