package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// ClusterClient creates a client for interfacing with the Instaclustr Cluster API
type ClusterClient struct {
	client *InstaclustrClient
}

// CreateClusterRequest is the request object for provisioning new clusters
type CreateClusterRequest struct {
	ClusterName string                     `json:"clusterName"`
	Provider    string                     `json:"provider"`
	Account     string                     `json:"account,omitempty"`
	Version     string                     `json:"version"`
	Size        string                     `json:"size"`
	Region      CreateClusterRequestRegion `json:"region"`
	Tags        map[string]string          `json:"tags,omitempty"`
}

// CreateClusterRequestRegion is the region sub section for cluster creation
type CreateClusterRequestRegion struct {
	Datacenter                    string                                     `json:"dataCentre"`
	AuthnAuthz                    bool                                       `json:"authnAuthz,string"`
	ClientEncryption              bool                                       `json:"clientEncryption,string"`
	UsePrivateBroadcastRPCAddress bool                                       `json:"usePrivateBroadcastRPCAddress,string"`
	DefaultNetwork                string                                     `json:"defaultNetwork"`
	FirewallRules                 []string                                   `json:"firewallRules"`
	RackAllocations               []CreateClusterRequestRegionRackAllocation `json:"rackAllocation"`
	DiskEncryptionKey             string                                     `json:"diskEncryptionKey,omitempty"`
}

// CreateClusterRequestRegionRackAllocation specifies rack allocation when provisioning a cluster
type CreateClusterRequestRegionRackAllocation struct {
	Name      string `json:"name"`
	NodeCount int    `json:"nodeCount,string"`
}

// CreateClusterResponse is the response from provisioning a cluster
type CreateClusterResponse struct {
	ID string `json:"id"`
}

// ClusterStatus is the returned object for a cluster
type ClusterStatus struct {
	ID                         string         `json:"id"`
	ClusterName                string         `json:"clusterName"`
	ClusterNetwork             ClusterNetwork `json:"clusterNetwork"`
	ClusterStatus              string         `json:"clusterStatus"`
	CassandraVersion           string         `json:"cassandraVersion"`
	Username                   string         `json:"username"`
	InstaclustrUserPassword    string         `json:"instaclustrUserPassword"`
	ClusterCertificateDownload string         `json:"clusterCertificateDownload"`
	Datacenters                []Datacenter   `json:"dataCentres"`
}

// ClusterNetwork is the network object for cluster
type ClusterNetwork struct {
	Network      string `json:"network"`
	PrefixLength int    `json:"prefixLength"`
}

// Datacenter is the datacenter object for a cluster
type Datacenter struct {
	ID                            string           `json:"id"`
	Name                          string           `json:"name"`
	Provider                      string           `json:"provider"`
	ClientEncryption              bool             `json:"clientEncryption"`
	PasswordAuthentication        bool             `json:"passwordAuthentication"`
	UserAuthorization             bool             `json:"userAuthorization"`
	UsePrivateBroadcastRPCAddress bool             `json:"usePrivateBroadcastRPCAddress"`
	CdcNetwork                    ClusterNetwork   `json:"cdcNetwork"`
	Bundles                       []string         `json:"bundles"`
	Nodes                         []DatacenterNode `json:"nodes"`
	NodeCount                     int              `json:"nodeCount"`
}

// DatacenterNode is the datacenter node object for a cluster datacenter
type DatacenterNode struct {
	ID             string `json:"id"`
	Size           string `json:"size"`
	Rack           string `json:"rack"`
	PublicAddress  string `json:"publicAddress"`
	PrivateAddress string `json:"privateAddress"`
	NodeStatus     string `json:"nodeStatus"`
	SparkMaster    bool   `json:"sparkMaster"`
	SparkJobserver bool   `json:"sparkJobserver"`
	Zeppelin       bool   `json:"zeppelin"`
}

// ClusterListStatus is the object returned by the cluster LIST endpoint
type ClusterListStatus struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	CassandraVersion string `json:"cassandraVersion"`
	NodeCount        int    `json:"nodeCount"`
	RunningNodeCount int    `json:"runningNodeCount"`
	DerivedStatus    string `json:"derivedStatus"`
}

// List returns a list of cluster statuses
func (c *ClusterClient) List() ([]*ClusterListStatus, error) {
	response, err := c.client.doGet("")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseData, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 || response.StatusCode != 202 {
		return nil, fmt.Errorf("List Cluster did not return 200/202 [%d]:\n%s", response.StatusCode, string(responseData))
	}
	clusters := []*ClusterListStatus{}
	err = json.Unmarshal(responseData, &clusters)
	if err != nil {
		return nil, err
	}
	return clusters, nil
}

// Get returns the status for a specific cluster
func (c *ClusterClient) Get(clusterID string) (*ClusterStatus, error) {
	response, err := c.client.doGet(clusterID)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseData, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 || response.StatusCode != 202 {
		return nil, fmt.Errorf("Get Cluster did not return 200/202 [%d]:\n%s", response.StatusCode, string(responseData))
	}
	cluster := &ClusterStatus{}
	err = json.Unmarshal(responseData, cluster)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

// Delete deletes a cluster
func (c *ClusterClient) Delete(clusterID string) error {
	response, err := c.client.doDelete(clusterID, nil)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 || response.StatusCode != 202 {
		return fmt.Errorf("Cluster DELETE did not return 200/202 [%d]", response.StatusCode)
	}
	return nil
}

// Create creates a new cluster
func (c *ClusterClient) Create(request CreateClusterRequest) (*CreateClusterResponse, error) {
	bytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	response, err := c.client.doPost("", bytes)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseData, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 || response.StatusCode != 202 {
		return nil, fmt.Errorf("Create Cluster did not return 200/202 [%d]:\n%s", response.StatusCode, string(responseData))
	}
	cluster := &CreateClusterResponse{}
	err = json.Unmarshal(responseData, cluster)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}
