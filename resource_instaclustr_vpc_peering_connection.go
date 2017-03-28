package main

import "github.com/hashicorp/terraform/helper/schema"

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
	return nil
}

func resourceInstaclustrVpcPeeringConnectionRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceInstaclustrVpcPeeringConnectionDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
