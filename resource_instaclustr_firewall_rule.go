package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstaclustrFirewallRuleCreate,
		Read:   resourceInstaclustrFirewallRuleRead,
		Update: resourceInstaclustrFirewallRuleUpdate,
		Delete: resourceInstaclustrFirewallRuleDelete,

		Schema: map[string]*schema.Schema{
			"rule": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceInstaclustrFirewallRuleCreate(d *schema.ResourceData, m interface{}) error {
	rule := d.Get("rule").(string)
	p := getJsonObject(d, m)

	if !isExistingFirewallRule(p, rule) {
		log.Println("[DEBUG] Creating firewall rule in instaclustr")
		addFireWallRule(rule, d, m)
		d.SetId("Instaclustr_" + rule)
	} else {
		log.Println("[DEBUG] Firewall rule was already set in instaclustr")
		d.SetId("Instaclustr_" + rule)
	}
	return nil
}

func resourceInstaclustrFirewallRuleRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceInstaclustrFirewallRuleUpdate(d *schema.ResourceData, m interface{}) error {
	oA, nA := d.GetChange("rule")
	oldrule := oA.(string)
	newrule := nA.(string)
	p := getJsonObject(d, m)

	if isExistingFirewallRule(p, oldrule) {
		log.Println("[DEBUG] Deleting firewall rule in instaclustr")
		deleteFireWallRule(oldrule, d, m)
	} else {
		log.Println("[DEBUG] Firewall rule was already deleted in instaclustr")
	}

	if !isExistingFirewallRule(p, newrule) {
		log.Println("[DEBUG] Creating firewall rule in instaclustr")
		addFireWallRule(newrule, d, m)
		d.SetId("Instaclustr_" + newrule)
	} else {
		log.Println("[DEBUG] Firewall rule was already set in instaclustr")
	}
	return nil
}

func resourceInstaclustrFirewallRuleDelete(d *schema.ResourceData, m interface{}) error {
	rule := d.Get("rule").(string)
	p := getJsonObject(d, m)

	if isExistingFirewallRule(p, rule) {
		log.Println("[DEBUG] Deleting firewall rule in instaclustr")
		deleteFireWallRule(rule, d, m)
		d.SetId("")
	} else {
		log.Println("[DEBUG] Firewall rule was already deleted in instaclustr")
	}
	return nil
}

type InstaclustrJSONData struct {
	Network string `json:"network"`
	Rules   []struct {
		Type   string `json:"type"`
		Status string `json:"status"`
	} `json:"rules"`
}

type InstaclustrJSON []InstaclustrJSONData

func addFireWallRule(rule string, d *schema.ResourceData, m interface{}) {
	evaluateFireWallRule("POST", rule, d, m)
}

func deleteFireWallRule(rule string, d *schema.ResourceData, m interface{}) {
	evaluateFireWallRule("DELETE", rule, d, m)
}

func evaluateFireWallRule(requestType string, rule string, d *schema.ResourceData, m interface{}) {
	cluster_id := d.Get("cluster_id").(string)
	config := m.(*Config)
	username := config.AccessKey
	passwd := config.SecretKey
	apiUrl := config.Url

	if username == "" {
		panic("Must set environment variable for instaclustr <INSTACLUSTR_ACCESS_KEY>")
	}
	if passwd == "" {
		panic("Must set environment variable for instaclustr <INSTACLUSTR_SECRET_KEY>")
	}

	var jsonStr = []byte(fmt.Sprintf(`
			{
				"network": "%s",
				"rules":[
					{
						"type":"CASSANDRA"
					}
				]
			}`, rule))

	req, err := http.NewRequest(requestType, strings.Join([]string{apiUrl, cluster_id, "firewallRules"}, "/"), bytes.NewBuffer(jsonStr))
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
	if err != nil {
		log.Fatal(err)
	}
	return response
}

func getRequestToString(resp *http.Response) string {
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(bodyText)
}
