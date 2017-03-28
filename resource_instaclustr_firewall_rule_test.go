package main

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccInstaclustrFirewallRule_basic(t *testing.T) {
	var firewall Firewall
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstaclustrFirewallRuleDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstaclustrFirewallRuleConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstaclustrFirewallRuleExists("instaclustr_firewall_rule.foo", &firewall),
					resource.TestCheckResourceAttr("instaclustr_firewall_rule.foo", "network", "172.23.0.0/22"),
				),
			},
		},
	})
}

func testAccCheckInstaclustrFirewallRuleExists(n string, fw *Firewall) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("Firewall Rule does not exists in state")
		}
		client := testAccProvider.Meta().(*InstaclustrClient).FirewallClient()
		firewalls, err := client.List(rs.Primary.Attributes["cluster_id"])
		if err != nil {
			return err
		}
		firewall := ruleForNetwork(rs.Primary.Attributes["network"], firewalls)
		if firewall == nil {
			return fmt.Errorf("Firewall Rule not found")
		}
		*fw = *firewall
		return nil
	}
}

func testAccCheckInstaclustrFirewallRuleDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*InstaclustrClient).FirewallClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "instaclustr_firewall_rule" {
			continue
		}
		firewalls, err := client.List(rs.Primary.Attributes["cluster_id"])
		if err == nil {
			firewall := ruleForNetwork(rs.Primary.Attributes["network"], firewalls)
			if firewall != nil {
				return fmt.Errorf("Firewall Rule still exists")
			}
		}
		return nil
	}
	return nil
}

func ruleForNetwork(network string, firewalls []*Firewall) *Firewall {
	for _, f := range firewalls {
		if f.Network == network {
			return f
		}
	}
	return nil
}

const testAccInstaclustrFirewallRuleConfig = `
resource "instaclustr_cluster" "foo" {
  name = "terraform-test-acc"
  account = "PeopleNet"
  provider_name = "AWS_VPC"
  version = "apache-cassandra-3.0.10"
  size = "t2.small"
  region_datacenter = "US_EAST_1"
  region_default_network = "10.0.0.0/16"
  region_rack_allocation {
    name = "us-east-1a"
    node_count = 1
  }
  region_rack_allocation {
    name = "us-east-1b"
    node_count = 1
  }
}

resource "instaclustr_firewall_rule" "foo" {
  cluster_id = "${instaclustr_cluster.foo.id}"
  network = "172.23.0.0/22"
}
`
