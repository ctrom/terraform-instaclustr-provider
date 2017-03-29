package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func testAccVpcPeeringCheckPreCheck(t *testing.T) {
	if v := os.Getenv("AWS_ACCESS_KEY_ID"); v == "" {
		t.Fatal("AWS_ACCESS_KEY_ID must be set for acceptance tests")
	}
	if v := os.Getenv("AWS_SECRET_ACCESS_KEY"); v == "" {
		t.Fatal("AWS_SECRET_ACCESS_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("AWS_REGION"); v == "" {
		t.Fatal("AWS_REGION must be set for acceptance tests")
	}
}

func TestAccInstaclustrVpcPeeringConnection_basic(t *testing.T) {
	var peeringConnection VpcPeer
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccVpcPeeringCheckPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstaclustrVpcPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstaclustrVpcPeeringConnectionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstaclustrVpcPeeringConnectionExists("instaclustr_vpc_peering_connection.main", &peeringConnection),
					resource.TestCheckResourceAttr("instaclustr_vpc_peering_connection.main", "status", "pending-acceptance"),
				),
			},
		},
	})
}

func testAccCheckInstaclustrVpcPeeringConnectionExists(n string, pc *VpcPeer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("VPC Peering Connection does not exists in state")
		}
		client := testAccProvider.Meta().(*InstaclustrClient).VpcPeeringClient()
		connection, err := client.Get(rs.Primary.Attributes["cluster_datacenter_id"], rs.Primary.ID)
		if err != nil {
			return err
		}
		if connection == nil {
			return fmt.Errorf("VPC Peering Connection not found")
		}
		*pc = *connection
		return nil
	}
}

func testAccCheckInstaclustrVpcPeeringConnectionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*InstaclustrClient).VpcPeeringClient()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "instaclustr_vpc_peering_connection" {
			continue
		}
		connection, err := client.Get(rs.Primary.Attributes["cluster_datacenter_id"], rs.Primary.ID)
		if err == nil && connection != nil {
			return fmt.Errorf("VPC Peering Connecting still exists")
		}
	}
	return nil
}

const testAccInstaclustrVpcPeeringConnectionConfig = `
resource "aws_vpc" "main" {
  cidr_block = "10.1.0.0/16"
}

data "aws_caller_identity" "current" {}

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

resource "instaclustr_vpc_peering_connection" "main" {
  peer_vpc_id = "${aws_vpc.main.id}"
  peer_account_id = "${data.aws_caller_identity.current.account_id}"
  peer_subnet = "${aws_vpc.main.cidr_block}"
  cluster_datacenter_id = "${instaclustr_cluster.foo.datacenter.0.datacenter_id}"
}
`
