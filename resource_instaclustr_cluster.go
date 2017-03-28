package main

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstaclustrClusterCreate,
		Read:   resourceInstaclustrClusterRead,
		Delete: resourceInstaclustrClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"provider_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"account": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"version": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"size": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region_datacenter": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region_auth": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"region_client_encryption": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"region_use_private_rpc_broadcast_address": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"region_default_network": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region_rack_allocation": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"node_count": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},
	}
}

func resourceInstaclustrClusterCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*InstaclustrClient).ClusterClient()
	request := CreateClusterRequest{
		ClusterName: d.Get("name").(string),
		Provider:    d.Get("provider_name").(string),
		Version:     d.Get("version").(string),
		Size:        d.Get("size").(string),
		Region: CreateClusterRequestRegion{
			Datacenter:                    d.Get("region_datacenter").(string),
			ClientEncryption:              d.Get("region_client_encryption").(bool),
			UsePrivateBroadcastRPCAddress: d.Get("region_use_private_rpc_broadcast_address").(bool),
			DefaultNetwork:                d.Get("region_default_network").(string),
			RackAllocations:               []CreateClusterRequestRegionRackAllocation{},
			FirewallRules:                 []string{},
		},
	}
	if account, ok := d.GetOk("account"); ok {
		request.Account = account.(string)
	}
	if auth, ok := d.GetOk("region_auth"); ok {
		request.Region.AuthnAuthz = auth.(bool)
	}
	for _, rack := range d.Get("region_rack_allocation").(*schema.Set).List() {
		alloc := rack.(map[string]interface{})
		request.Region.RackAllocations = append(request.Region.RackAllocations, CreateClusterRequestRegionRackAllocation{
			Name:      alloc["name"].(string),
			NodeCount: alloc["node_count"].(int),
		})
	}
	response, err := client.Create(request)
	if err != nil {
		return err
	}
	d.SetId(response.ID)
	return resourceInstaclustrClusterRead(d, m)
}

func resourceInstaclustrClusterRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*InstaclustrClient).ClusterClient()
	cluster, err := client.Get(d.Id())
	if err != nil {
		d.SetId("")
		return err
	}
	d.Set("name", cluster.ClusterName)
	d.Set("version", cluster.CassandraVersion)
	d.Set("region_default_network", fmt.Sprintf("%s/%d", cluster.ClusterNetwork.Network, cluster.ClusterNetwork.PrefixLength))

	datacenter := cluster.Datacenters[0]
	d.Set("provider_name", datacenter.Provider)
	d.Set("region_datacenter", datacenter.Name)
	d.Set("region_auth", datacenter.PasswordAuthentication && datacenter.UserAuthorization)
	d.Set("region_client_encryption", datacenter.ClientEncryption)
	d.Set("region_use_private_rpc_broadcast_address", datacenter.UsePrivateBroadcastRPCAddress)

	racks := map[string]map[string]interface{}{}
	for _, n := range datacenter.Nodes {
		rack := racks[n.Rack]
		if rack == nil {
			rack = map[string]interface{}{
				"name":       n.Rack,
				"node_count": 0,
			}
			racks[n.Rack] = rack
		}
		rack["node_count"] = rack["node_count"].(int) + 1
	}
	rackSet := []interface{}{}
	for _, rack := range racks {
		rackSet = append(rackSet, rack)
	}
	d.Set("rack_allocation", schema.NewSet(rackHash, rackSet))

	// From node
	node := datacenter.Nodes[0]
	d.Set("size", node.Size)

	return nil
}

func resourceInstaclustrClusterDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*InstaclustrClient).ClusterClient()
	err := client.Delete(d.Id())
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func rackHash(v interface{}) int {
	var buf bytes.Buffer
	a := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-%d", a["name"].(string), a["node_count"].(int)))
	return hashcode.String(buf.String())
}
