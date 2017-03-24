package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"net/http"
	"log"
	"io/ioutil"
	"fmt"
	"bytes"
	"encoding/json"
)

func firewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCreate,
		Read:   resourceServerRead,
		Update: resourceServerUpdate,
		Delete: resourceServerDelete,

		Schema: map[string]*schema.Schema{
			"address": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"firewall_rules_url": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceServerCreate(d *schema.ResourceData, m interface{}) error {
	address := d.Get("address").(string)
	p := getJsonObject(d, m)

	if !isExistingFirewallRule(p, address) {
		log.Println("[DEBUG] Creating firewall rule in instaclustr")
		addFireWallRule(address, d, m)
		d.SetId("Instaclustr_" + address)
	} else {
		log.Println("[DEBUG] Firewall rule was already set in instaclustr")
		d.SetId("Instaclustr_" + address)
	}
	return nil
}

func resourceServerRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceServerUpdate(d *schema.ResourceData, m interface{}) error {
	oA, nA := d.GetChange("address")
	oldAddress := oA.(string)
	newAddress := nA.(string)
	p := getJsonObject(d, m)

	if isExistingFirewallRule(p, oldAddress) {
		log.Println("[DEBUG] Deleting firewall rule in instaclustr")
		deleteFireWallRule(oldAddress, d, m)
	} else {
		log.Println("[DEBUG] Firewall rule was already deleted in instaclustr")
	}

	if !isExistingFirewallRule(p, newAddress) {
		log.Println("[DEBUG] Creating firewall rule in instaclustr")
		addFireWallRule(newAddress, d, m)
		d.SetId("Instaclustr_" + newAddress)
	} else {
		log.Println("[DEBUG] Firewall rule was already set in instaclustr")
	}
	return nil
}

func resourceServerDelete(d *schema.ResourceData, m interface{}) error {
	address := d.Get("address").(string)
	p := getJsonObject(d, m)

	if isExistingFirewallRule(p, address) {
		log.Println("[DEBUG] Deleting firewall rule in instaclustr")
		deleteFireWallRule(address, d, m)
		d.SetId("")
	} else {
		log.Println("[DEBUG] Firewall rule was already deleted in instaclustr")
	}
	return nil
}

type InstaclustrJSONData struct {
	Network string `json:"network"`
	Rules []struct {
		Type string `json:"type"`
		Status string `json:"status"`
	} `json:"rules"`
}

type InstaclustrJSON []InstaclustrJSONData

func addFireWallRule(address string, d *schema.ResourceData, m interface{}) {
	evaluateFireWallRule("POST", address, d, m)
}

func deleteFireWallRule(address string, d *schema.ResourceData, m interface{}) {
	evaluateFireWallRule("DELETE", address, d, m)
}

func evaluateFireWallRule(requestType string, address string, d *schema.ResourceData, m interface{}) {
	firewall_rules_url := d.Get("firewall_rules_url").(string)
	config := m.(*Config)
	username := config.AccessKey
	passwd := config.SecretKey

	if username == "" {
		panic("Must set environment variable for instaclustr <INSTACLUSTR_ACCESS_KEY>")
	}
	if passwd == "" {
		panic("Must set environment variable for instaclustr <INSTACLUSTR_SECRET_KEY>")
	}

	var jsonStr = []byte(fmt.Sprintf(`
			{
				"network":"%s",
				"rules":[
					{
						"type":"CASSANDRA"
					}
				]
			}`, address))

	req, err := http.NewRequest(requestType, firewall_rules_url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "testing")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, passwd)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func isExistingFirewallRule(p InstaclustrJSON, ip string) bool {
	firewallRules := len(p)
	firewallRulesExist := false
	for i := 0; i < firewallRules; i++ {
		if p[i].Network == ip {
			firewallRulesExist = true
		}
	}
	return firewallRulesExist
}

func getJsonObject(d *schema.ResourceData, m interface{}) InstaclustrJSON {
	var p InstaclustrJSON
	s := getJsonText(d, m)

	errs := json.Unmarshal([]byte(s), &p)
	if errs != nil {
		panic(errs)
	}
	return p
}

func getJsonText(d *schema.ResourceData, m interface{}) string {
	resp := getRequest(d, m)
	return getRequestToString(resp)
}

func getRequest(d *schema.ResourceData, m interface{}) *http.Response {
	config := m.(*Config)
	username := config.AccessKey
	passwd := config.SecretKey
	if username == "" {
		panic("Must set environment variable for instaclustr <INSTACLUSTR_ACCESS_KEY>")
	}
	if passwd == "" {
		panic("Must set environment variable for instaclustr <INSTACLUSTR_SECRET_KEY>")
	}
	client := &http.Client{}
	firewall_rules_url := d.Get("firewall_rules_url").(string)
	request, err := http.NewRequest("GET", firewall_rules_url, nil)

	request.SetBasicAuth(username, passwd)
	response, err := client.Do(request)
	if err != nil{
		log.Fatal(err)
	}
	return response
}

func getRequestToString(resp *http.Response) string {
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		log.Fatal(err)
	}
	return string(bodyText)
}