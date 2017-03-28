package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

// FirewallClient is a client for interacting with Firewall rules
type FirewallClient struct {
	client *InstaclustrClient
}

// Firewall is the response from the Firewall API
type Firewall struct {
	Network string         `json:"network"`
	Rules   []FirewallRule `json:"rules"`
}

// FirewallRule is a collection inside Firewall
type FirewallRule struct {
	Type string `json:"type"`
}

// List returns the firewall rules for a cluster
func (fc *FirewallClient) List(clusterID string) ([]Firewall, error) {
	response, err := fc.client.doGet(strings.Join([]string{clusterID, "firewallRules"}, "/"))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	firewall := []Firewall{}
	err = json.NewDecoder(response.Body).Decode(firewall)
	if err != nil {
		return nil, err
	}
	return firewall, nil
}

// Create adds a firewall rule to a cluster for the provided network CIDR
func (fc *FirewallClient) Create(clusterID, network string) error {
	firewall := Firewall{
		Network: network,
		Rules: []FirewallRule{
			FirewallRule{
				Type: "CASSANDRA",
			},
			FirewallRule{
				Type: "SPARK",
			},
			FirewallRule{
				Type: "SPARK_JOBSERVER",
			},
		},
	}
	bytes, err := json.Marshal(firewall)
	if err != nil {
		return err
	}
	response, err := fc.client.doPost(strings.Join([]string{clusterID, "firewallRules"}, "/"), bytes)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return nil
}

// Delete removes a network firewall rule from a cluster
func (fc *FirewallClient) Delete(clusterID, network string) error {
	firewall := Firewall{
		Network: network,
		Rules: []FirewallRule{
			FirewallRule{
				Type: "CASSANDRA",
			},
			FirewallRule{
				Type: "SPARK",
			},
			FirewallRule{
				Type: "SPARK_JOBSERVER",
			},
		},
	}
	bytes, err := json.Marshal(firewall)
	if err != nil {
		return err
	}
	response, err := fc.client.doDelete(strings.Join([]string{clusterID, "firewallRules"}, "/"), bytes)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return nil
}

func resourceFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstaclustrFirewallRuleCreate,
		Read:   resourceInstaclustrFirewallRuleRead,
		Delete: resourceInstaclustrFirewallRuleDelete,

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
			networkRule = &f
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

func firewallID(clusterID, network string) string {
	return fmt.Sprintf("%s:%s", clusterID, network)
}
