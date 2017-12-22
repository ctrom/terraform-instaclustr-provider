# A terraform provider for instaclustr.  

Built using API defined at https://support.instaclustr.com/hc/en-us/articles/213671118-Provisioning-API

```
provider "instaclustr" "test"{
  access_key = "username" // will automatically use INSTACLUSTR_ACCESS_KEY envvar
  secret_key = "API key" // will automatically use INSTACLUSTR_SECRET_KEY envvar
  //url = "Override the API URL if desired"
}
```

## Resources

### Cluster

```
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
```

#### Arguments

* `name` - the cluster's name
* `version` - the cluster's cassandra version. Obtain values from Instaclustr dashboard
* `datacenter` - Defines a datacenter for the cluster. Currently can only provide 1
  * `provider_name` - the provider for the datacenter. One of: `AWS_VPC`, `AZURE`, `SOFTLAYER_BARE_METAL`, `GCP`
  * `account` - (Optional) the account name for provisioning resources. Obtain from Instaclustr dashboard.
  * `region` - The region to deploy the datacenter in. Provider specific. See API docs.
  * `size` - The node instance sizes. Provider specific. See API docs.
  * `auth` - (Optional) Enables authentication for the datacenter. Default `false`
  * `disk_encryption_key` - (Optional) UUID of KMS key in AWS. Enables client encryption when provided. Not supported on T2 instances. Default `""`
  * `use_private_rpc_broadcast_address` - (Optional) use the private IP address for cluster communication. Default `true`.
  * `default_network` - The CIDR network for the datacenter.
  * `rack` - Defines a server rack for the datacenter. Must define at minimum 2
    * `name` - The rack name
    * `node_count` - The number of instances in the rack

#### Attributes

* `private_ips` - list of node private IP addresses
* `public_ips` - list of node public IP addresses
* `datacenter`
  * `datacenter_id` - the ID of the datacenter

### Firewall Rule

```
resource "instaclustr_firewall_rule" "foo" {
  cluster_id = "${instaclustr_cluster.foo.id}"
  network = "10.1.0.0/16"
}
```

#### Arguments

* `cluster_id` - the cluster ID to add the firewall rule to
* `network` - the network CIDR block to authorize for access

#### Attributes

* none

### VPC Peering Connection

```
resource "instaclustr_vpc_peering_connection" "main" {
  peer_vpc_id = "${aws_vpc.main.id}"
  peer_account_id = "${data.aws_caller_identity.current.account_id}"
  peer_subnet = "${aws_vpc.main.cidr_block}"
  cluster_datacenter_id = "${instaclustr_cluster.foo.datacenter.0.datacenter_id}"
}
```

#### Arguments

* `peer_vpc_id` - the ID of the VPC to peer to the cluster's datacenter. Must be in the same region.
* `peer_account_id` - the AWS account ID for the VPC to peer with
* `peer_subnet` - the network CIDR for the VPC to peer with
* `cluster_datacenter_id` - the ID of the cluster's datacenter to create the connection

#### Attributes

* `vpc_id` - the vpc ID of the cluster's datacenter in AWS
* `aws_vpc_connection_id` - the ID of the vpc peering connection
* `status` - the status of the VPC peering connection

## Datasources

### Cluster IPs

```
data "instaclustr_cluster_ips" "cluster" {
  cluster_id = "${instaclustr_cluster.foo.id}"
}
```

#### Arguments

* `cluster_id` - the cluster ID to collect IPs for

#### Attributes

* `private_ips` - list of node private IP addresses
* `public_ips` - list of node public IP addresses
* `cidr_block` - the CIDR block for the cluster's datacenter (only the first entry)

