package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
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
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"initiating-request", "pending-acceptance", "provisioning", "active"},
		Target:     []string{"pending-acceptance", "active"},
		Refresh:    vpcConnectionStateRefreshFunc(client, clusterDatacenterID, response.ID),
		Timeout:    15 * time.Minute,
		Delay:      3 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, waitErr := stateConf.WaitForState()
	if waitErr != nil {
		return fmt.Errorf(
			"Error waiting for VPC Peering Connecting (%s) to be ready: %s", response.ID, waitErr)
	}
	d.SetId(vpcPeeringConnectionID(clusterDatacenterID, response.ID))
	return resourceInstaclustrVpcPeeringConnectionRead(d, m)
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
	d.Set("peer_subnet", vpcPeer.PeerSubnet)
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

func vpcConnectionStateRefreshFunc(client *VpcPeeringClient, datacenterID, connectionID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		connection, err := client.Get(datacenterID, connectionID)
		if err != nil {
			return nil, "", err
		}
		return connection, connection.StatusCode, nil
	}
}
