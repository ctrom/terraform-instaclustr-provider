package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
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
			"version": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"public_ips": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"private_ips": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"datacenter": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provider_name": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: stringInList([]string{"AWS_VPC", "AZURE", "SOFTLAYER_BARE_METAL", "GCP"}),
						},
						"account": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"region": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"size": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"datacenter_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"auth": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							ForceNew: true,
						},
						"client_encryption": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
						"disk_encryption_key": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"use_private_rpc_broadcast_address": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"default_network": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"rack": &schema.Schema{
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
				},
			},
		},
	}
}

func resourceInstaclustrClusterCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*InstaclustrClient).ClusterClient()
	datacenter := d.Get("datacenter").([]interface{})[0].(map[string]interface{})
	request := CreateClusterRequest{
		ClusterName: d.Get("name").(string),
		Version:     d.Get("version").(string),

		Provider: datacenter["provider_name"].(string),
		Size:     datacenter["size"].(string),
		Region: CreateClusterRequestRegion{
			Datacenter:                    datacenter["region"].(string),
			UsePrivateBroadcastRPCAddress: datacenter["use_private_rpc_broadcast_address"].(bool),
			DefaultNetwork:                datacenter["default_network"].(string),
			AuthnAuthz:                    datacenter["auth"].(bool),
			ClientEncryption:              false,
			RackAllocations:               []CreateClusterRequestRegionRackAllocation{},
			FirewallRules:                 []string{},
		},
	}
	if account, ok := datacenter["account"]; ok {
		request.Account = account.(string)
	}
	if key, ok := datacenter["disk_encryption_key"]; ok && key != "" {
		request.Region.ClientEncryption = true
		request.Region.DiskEncryptionKey = key.(string)
	}

	for _, rack := range datacenter["rack"].(*schema.Set).List() {
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
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"RUNNING", "GENESIS", "PROVISIONING", "PROVISIONED"},
		Target:     []string{"RUNNING"},
		Refresh:    clusterStateRefreshFunc(client, response.ID),
		Timeout:    15 * time.Minute,
		Delay:      3 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, waitErr := stateConf.WaitForState()
	if waitErr != nil {
		return fmt.Errorf(
			"Error waiting for Cluster (%s) to be Running: %s", response.ID, waitErr)
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

	//Set non-Computed values first
	d.Set("name", cluster.ClusterName)
	d.Set("version", cluster.CassandraVersion)

	datacenter := cluster.Datacenters[0]
	node := datacenter.Nodes[0]
	dcResource := d.Get("datacenter")
	dc := dcResource.([]interface{})[0].(map[string]interface{})
	dc["provider_name"] = datacenter.Provider
	dc["region"] = datacenter.Name
	dc["size"] = node.Size
	dc["auth"] = datacenter.PasswordAuthentication && datacenter.UserAuthorization
	dc["use_private_rpc_broadcast_address"] = datacenter.UsePrivateBroadcastRPCAddress
	dc["default_network"] = fmt.Sprintf("%s/%d", cluster.ClusterNetwork.Network, cluster.ClusterNetwork.PrefixLength)

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
	dc["rack"] = rackSet
	d.Set("datacenter", dcResource)

	// Set computed values last after reaquiring the datacenter map
	dcResource = d.Get("datacenter")
	dc = dcResource.([]interface{})[0].(map[string]interface{})
	dc["datacenter_id"] = datacenter.ID
	dc["client_encryption"] = datacenter.ClientEncryption
	d.Set("datacenter", dcResource)

	publicIps, privateIps := ipsForCluster(cluster)
	d.Set("public_ips", publicIps)
	d.Set("private_ips", privateIps)
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

func datacenterHash(v interface{}) int {
	var buf bytes.Buffer
	datacenter := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", datacenter["provider_name"].(string)))
	if account, ok := datacenter["account"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", account.(string)))
	}
	buf.WriteString(fmt.Sprintf("%s-", datacenter["region"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", datacenter["size"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", datacenter["datacenter_id"].(string)))
	buf.WriteString(fmt.Sprintf("%t-", datacenter["auth"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", datacenter["client_encryption"].(bool)))
	buf.WriteString(fmt.Sprintf("%t-", datacenter["use_private_rpc_broadcast_address"].(bool)))
	buf.WriteString(fmt.Sprintf("%s-", datacenter["default_network"].(string)))

	for _, rack := range datacenter["rack"].([]interface{}) {
		var buf2 bytes.Buffer
		r := rack.(map[string]interface{})
		buf2.WriteString(fmt.Sprintf("%s-%d", r["name"], r["node_count"]))
		buf.WriteString(fmt.Sprintf("%d-", hashcode.String(buf2.String())))
	}
	return hashcode.String(buf.String())
}

func clusterStateRefreshFunc(client *ClusterClient, clusterID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := client.Get(clusterID)
		if err != nil {
			return nil, "", err
		}
		return cluster, cluster.ClusterStatus, nil
	}
}
