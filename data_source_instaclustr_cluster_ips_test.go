package main

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccInstaclustrClusterIpsDatasource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstaclustrClusterIpsDatasourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.instaclustr_cluster_ips.ips", "private_ips"),
				),
			},
		},
	})
}

const testAccInstaclustrClusterIpsDatasourceConfig = `
resource "instaclustr_cluster" "foo" {
  name = "terraform-test-acc"
  version = "apache-cassandra-3.0.10"
  datacenter {
    provider_name = "AWS_VPC"
    account = "PeopleNet"
    region = "US_EAST_1"
    size = "t2.small"
    default_network = "10.0.0.0/16"
    rack {
      name = "us-east-1a"
      node_count = 1
    }
    rack {
      name = "us-east-1b"
      node_count = 1
    }
  }
}

data "instaclustr_cluster_ips" "ips" {
  cluster_id = "${instaclustr_cluster.foo.id}"
}
`
