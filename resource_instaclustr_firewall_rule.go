package main

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstaclustrFirewallRuleCreate,
		Read:   resourceInstaclustrFirewallRuleRead,
		Delete: resourceInstaclustrFirewallRuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceInstaclustrFirewallRuleImport,
		},

		Schema: map[string]*schema.Schema{
			"network": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceInstaclustrFirewallRuleCreate(d *schema.ResourceData, m interface{}) error {
	network := d.Get("network").(string)
	clusterID := d.Get("cluster_id").(string)
	client := m.(*InstaclustrClient).FirewallClient()
	err := client.Create(clusterID, network)
	if err != nil {
		d.SetId("")
		return err
	}
	return resourceInstaclustrFirewallRuleRead(d, m)
}

func resourceInstaclustrFirewallRuleRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*InstaclustrClient).FirewallClient()
	clusterID := d.Get("cluster_id").(string)
	firewallRules, err := client.List(clusterID)
	if err != nil {
		return err
	}
	var networkRule *Firewall
	for _, f := range firewallRules {
		if f.Network == d.Get("network").(string) {
			networkRule = f
		}
	}
	if networkRule == nil {
		d.SetId("")
	} else {
		d.SetId(firewallID(clusterID, networkRule.Network))
	}
	return nil
}

func resourceInstaclustrFirewallRuleDelete(d *schema.ResourceData, m interface{}) error {
	network := d.Get("network").(string)
	clusterID := d.Get("clusterId").(string)
	client := m.(*InstaclustrClient).FirewallClient()
	err := client.Delete(clusterID, network)
	if err != nil {
		return nil
	}
	d.SetId("")
	return nil
}

func resourceInstaclustrFirewallRuleImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	err := resourceInstaclustrFirewallRuleRead(d, m)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	return []*schema.ResourceData{d}, nil
}

func firewallID(clusterID, network string) string {
	return fmt.Sprintf("%s:%s", clusterID, network)
}
