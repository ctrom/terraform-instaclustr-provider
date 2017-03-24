package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INSTACLUSTR_ACCESS_KEY", ""),
				Description: "Instaclustr user name used for api access",
				Sensitive:   true,
			},
			"secret_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INSTACLUSTR_SECRET_KEY", ""),
				Description: "Instaclustr key used for api access",
				Sensitive:   true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"instaclustr_firewallrule": firewallRule(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {

	config := Config{
		AccessKey:    d.Get("access_key").(string),
		SecretKey:    d.Get("secret_key").(string),
	}

	return &config, nil
}