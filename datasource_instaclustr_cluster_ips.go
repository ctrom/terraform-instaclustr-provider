package main

import "github.com/hashicorp/terraform/helper/schema"

func dataSourceInstaclustrClusterIPs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceInstaclustrClusterIPsRead,

		Schema: map[string]*schema.Schema{
			"cluster_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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
		},
	}
}

func dataSourceInstaclustrClusterIPsRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*InstaclustrClient).ClusterClient()
	cluster, err := client.Get(d.Id())
	if err != nil {
		return err
	}
	publicIps, privateIps := ipsForCluster(cluster)

	d.Set("public_ips", publicIps)
	d.Set("private_ips", privateIps)

	return nil
}

func ipsForCluster(cluster *ClusterStatus) (publicIps []string, privateIps []string) {
	publicIps = []string{}
	privateIps = []string{}

	for _, datacenter := range cluster.Datacenters {
		for _, node := range datacenter.Nodes {
			privateIps = append(privateIps, node.PrivateAddress)
			if node.PublicAddress != "" {
				publicIps = append(publicIps, node.PublicAddress)
			}
		}
	}
	return publicIps, privateIps
}
