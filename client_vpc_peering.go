package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

// VpcPeeringClient provides an interface to the VPC Peering API
type VpcPeeringClient struct {
	client *InstaclustrClient
}

// VpcPeer is the response from the VPC Peering API
type VpcPeer struct {
	ID                  string        `json:"id"`
	AWSVpcConnectionID  string        `json:"aws_vpc_connection_id"`
	ClusterDatacenterID string        `json:"clusterDataCentre"`
	VpcID               string        `json:"vpcId"`
	PeerVpcID           string        `json:"peerVpcId"`
	PeerAccountID       string        `json:"peerAccountId"`
	PeerSubnet          VpcPeerSubnet `json:"peerSubnet"`
	StatusCode          string        `json:"statusCode"`
}

// VpcPeerSubnet is the subnet subsection for the peer configuration
type VpcPeerSubnet struct {
	Network      string `json:"network"`
	PrefixLength int    `json:"prefixLength"`
}

// CreateVpcPeerRequest is the object for creating a new VPC peering request
type CreateVpcPeerRequest struct {
	PeerVpcID     string `json:"peerVpcId"`
	PeerAccountID string `json:"peerAccountId"`
	PeerSubnet    string `json:"peerSubnet"`
}

// CreateVpcPeerResponse is the object returned when creating a new VPC peering request
type CreateVpcPeerResponse struct {
	ID string `json:"id"`
}

// List retrieves all VPC Peering connections for a cluster
func (c *VpcPeeringClient) List(clusterDatacenterID string) ([]*VpcPeer, error) {
	response, err := c.client.doGet(strings.Join([]string{"vpc-peering", clusterDatacenterID}, "/"))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseData, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 || response.StatusCode != 202 {
		return nil, fmt.Errorf("List VPC Peering Connection did not return 200/202 [%d]:\n%s", response.StatusCode, string(responseData))
	}
	vpcPeers := []*VpcPeer{}
	err = json.Unmarshal(responseData, &vpcPeers)
	if err != nil {
		return nil, err
	}
	return vpcPeers, nil
}

// Get retrieves the details for a single VPC Peering Connection on a cluster
func (c *VpcPeeringClient) Get(clusterDatacenterID, vpcPeeringConnectionID string) (*VpcPeer, error) {
	response, err := c.client.doGet(strings.Join([]string{"vpc-peering", clusterDatacenterID, vpcPeeringConnectionID}, "/"))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseData, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 || response.StatusCode != 202 {
		return nil, fmt.Errorf("Get VPC Peering Connection did not return 200/202 [%d]:\n%s", response.StatusCode, string(responseData))
	}
	vpcPeer := &VpcPeer{}
	err = json.Unmarshal(responseData, vpcPeer)
	if err != nil {
		return nil, err
	}
	return vpcPeer, nil
}

// Create submits a new VPC Peering Request
func (c *VpcPeeringClient) Create(clusterDatacenterID string, request *CreateVpcPeerRequest) (*CreateVpcPeerResponse, error) {
	bytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	response, err := c.client.doPost(strings.Join([]string{"vpc-peering", clusterDatacenterID}, "/"), bytes)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseData, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 || response.StatusCode != 202 {
		return nil, fmt.Errorf("Create VPC Peering Connection did not return 200/202 [%d]:\n%s\n%s\n%s", response.StatusCode, string(responseData), string(bytes), "ID: "+clusterDatacenterID)
	}
	createResponse := &CreateVpcPeerResponse{}
	err = json.Unmarshal(responseData, createResponse)
	if err != nil {
		return nil, err
	}
	return createResponse, nil
}

// Delete deletes an requested or existing VPC Peering Connection
func (c *VpcPeeringClient) Delete(clusterDatacenterID, vpcPeeringConnectionID string) error {
	response, err := c.client.doDelete(strings.Join([]string{"vpc-peering", clusterDatacenterID, vpcPeeringConnectionID}, "/"), nil)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	responseData, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 || response.StatusCode != 202 {
		return fmt.Errorf("Delete VPC Peering Connection did not return 200/202 [%d]:\n%s", response.StatusCode, string(responseData))
	}
	return nil
}
