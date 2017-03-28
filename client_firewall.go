package main

import (
	"encoding/json"
	"fmt"
	"strings"
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
func (fc *FirewallClient) List(clusterID string) ([]*Firewall, error) {
	response, err := fc.client.doGet(strings.Join([]string{clusterID, "firewallRules"}, "/"))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	firewall := []*Firewall{}
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
	if response.StatusCode != 202 {
		return fmt.Errorf("Firewall Delete did not return 202 [%d]", response.StatusCode)
	}
	return nil
}
