package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	responseData, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 && response.StatusCode != 202 {
		return nil, fmt.Errorf("List Firewall Rules did not return 200/202 [%d]:\n%s", response.StatusCode, string(responseData))
	}
	firewall := []*Firewall{}
	err = json.Unmarshal(responseData, &firewall)
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
	responseData, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 && response.StatusCode != 202 {
		return fmt.Errorf("Create Firewall Rule did not return 200/202 [%d]:\n%s\n%s", response.StatusCode, string(responseData), string(bytes))
	}
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
	responseData, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 && response.StatusCode != 202 {
		return fmt.Errorf("Delete Firewall Rule did not return 200/202 [%d]:\n%s", response.StatusCode, string(responseData))
	}
	return nil
}
