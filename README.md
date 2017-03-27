# A terraform provider for instaclustr.  
Currently the plugin gives access to one resource to add firewall rules.  
An example .tf setup
```
provider "instaclustr" "test"{}

resource "instaclustr_firewallrule" "foo" {  
  rule = "[IP_ADDRESS_TO_ADD]>"  
  firewall_rules_url = "[YOUR_CLUSTER_ID]"  
}
```
To build this plugin run:  
go build -o terraform-provider-instaclustr

INSTACLUSTR_ACCESS_KEY and INSTACLUSTR_SECRET_KEY must be set as environment variables for instaclustr to have access to your target firewall_rules_url.

An example terraform command:  
INSTACLUSTR_ACCESS_KEY=[ACCESS_KEY] INSTACLUSTR_SECRET_KEY=[SECRET_KEY] terraform [COMMAND]
