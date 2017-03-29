package main

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccInstaclustrCluster_basic(t *testing.T) {
	var cluster ClusterStatus
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstaclustrClusterDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstaclustrClusterConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstaclustrClusterExists("instaclustr_cluster.foo", &cluster),
					resource.TestCheckResourceAttrSet("instaclustr_cluster.foo", "datacenter.0.datacenter_id"),
				),
			},
		},
	})
}

func testAccCheckInstaclustrClusterExists(n string, c *ClusterStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("Cluster does not exists in state")
		}
		client := testAccProvider.Meta().(*InstaclustrClient).ClusterClient()
		cluster, err := client.Get(rs.Primary.ID)
		if err != nil {
			return err
		}
		if cluster.ID != rs.Primary.ID {
			return fmt.Errorf("Cluster not found")
		}
		*c = *cluster
		return nil
	}
}

func testAccCheckInstaclustrClusterDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*InstaclustrClient).ClusterClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "instaclustr_cluster" {
			continue
		}
		cluster, err := client.Get(rs.Primary.ID)
		if err == nil {
			if cluster != nil && cluster.ID == rs.Primary.ID {
				return fmt.Errorf("Cluster still exists")
			}
		}
	}
	return nil
}

const testAccInstaclustrClusterConfig = `
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
`
