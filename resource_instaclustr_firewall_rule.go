package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstaclustrFirewallRuleCreate,
		Read:   resourceInstaclustrFirewallRuleRead,
		Delete: resourceInstaclustrFirewallRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
	client := m.(*InstaclustrClient).FirewallClient()
	network := d.Get("network").(string)
	clusterID := d.Get("cluster_id").(string)
	err := client.Create(clusterID, network)
	if err != nil {
		d.SetId("")
		return err
	}
	return resourceInstaclustrFirewallRuleRead(d, m)
}

func resourceInstaclustrFirewallRuleRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*InstaclustrClient).FirewallClient()
	clusterID, network, err := splitFirewallID(d.Id())
	if err != nil {
		d.SetId("")
		return err
	}
	firewallRules, err := client.List(clusterID)
	if err != nil {
		return err
	}
	var networkRule *Firewall
	for _, f := range firewallRules {
		if f.Network == network {
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
	client := m.(*InstaclustrClient).FirewallClient()
	clusterID, network, err := splitFirewallID(d.Id())
	if err != nil {
		d.SetId("")
		return err
	}
	err = client.Delete(clusterID, network)
	if err != nil {
		return nil
	}
	d.SetId("")
	return nil
}

func firewallID(clusterID, network string) string {
	return fmt.Sprintf("%s:%s", clusterID, network)
}

func splitFirewallID(id string) (string, string, error) {
	tokens := strings.Split(id, ":")
	if len(tokens) != 2 {
		return "", "", errors.New("Must supply ID in format of <clusterDatacenterID>:<network>")
	}
	return tokens[0], tokens[1], nil
}
