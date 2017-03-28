package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVpcPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstaclustrVpcPeeringConnectionCreate,
		Read:   resourceInstaclustrVpcPeeringConnectionRead,
		Delete: resourceInstaclustrVpcPeeringConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"peer_vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"peer_account_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"peer_subnet": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster_datacenter_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_vpc_connection_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceInstaclustrVpcPeeringConnectionCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*InstaclustrClient).VpcPeeringClient()
	clusterDatacenterID := d.Get("cluster_datacenter_id").(string)
	request := &CreateVpcPeerRequest{
		PeerAccountID: d.Get("peer_account_id").(string),
		PeerVpcID:     d.Get("peer_vpc_id").(string),
		PeerSubnet:    d.Get("peer_subnet").(string),
	}
	response, err := client.Create(clusterDatacenterID, request)
	if err != nil {
		d.SetId("")
		return err
	}
	d.SetId(vpcPeeringConnectionID(clusterDatacenterID, response.ID))
	return resourceInstaclustrFirewallRuleRead(d, m)
}

func resourceInstaclustrVpcPeeringConnectionRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*InstaclustrClient).VpcPeeringClient()
	clusterDatacenterID, id, err := splitVpcPeeringConnectionID(d.Id())
	if err != nil {
		d.SetId("")
		return err
	}
	vpcPeer, err := client.Get(clusterDatacenterID, id)
	if err != nil {
		d.SetId("")
		return err
	}
	d.Set("peer_vpc_id", vpcPeer.PeerVpcID)
	d.Set("peer_account_id", vpcPeer.PeerAccountID)
	d.Set("peer_subnet", fmt.Sprintf("%s/%d", vpcPeer.PeerSubnet.Network, vpcPeer.PeerSubnet.PrefixLength))
	d.Set("vpc_id", vpcPeer.VpcID)
	d.Set("aws_vpc_connection_id", vpcPeer.AWSVpcConnectionID)
	d.Set("status", vpcPeer.StatusCode)
	d.SetId(vpcPeeringConnectionID(vpcPeer.ClusterDatacenterID, vpcPeer.ID))
	return nil
}

func resourceInstaclustrVpcPeeringConnectionDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*InstaclustrClient).VpcPeeringClient()
	clusterDatacenterID, id, err := splitVpcPeeringConnectionID(d.Id())
	if err != nil {
		d.SetId("")
		return err
	}
	err = client.Delete(clusterDatacenterID, id)
	if err != nil {
		return nil
	}
	d.SetId("")
	return nil
}

func vpcPeeringConnectionID(clusterDatacenterID, vpcPeeringConnectionID string) string {
	return fmt.Sprintf("%s:%s", clusterDatacenterID, vpcPeeringConnectionID)
}

func splitVpcPeeringConnectionID(id string) (string, string, error) {
	tokens := strings.Split(id, ":")
	if len(tokens) != 2 {
		return "", "", errors.New("Must supply ID in format of <clusterDatacenterID>:<vpcPeeringConnectionID>")
	}
	return tokens[0], tokens[1], nil
}
