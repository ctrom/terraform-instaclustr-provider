package main

import (
	"bytes"
	"net/http"
	"strings"
)

// InstaclustrClient is a client for interfacing with the Instaclustr API
type InstaclustrClient struct {
	config Config
	client *http.Client
}

// FirewallClient creates a client for interfacing with the Instaclustr Firewall API
func (c *InstaclustrClient) FirewallClient() *FirewallClient {
	return &FirewallClient{
		client: c,
	}
}

// VpcPeeringClient creates a client fo rinterfacing with the Instaclustr Vpc Peering API
func (c *InstaclustrClient) VpcPeeringClient() *VpcPeeringClient {
	return &VpcPeeringClient{
		client: c,
	}
}

func (c *InstaclustrClient) doGet(path string) (*http.Response, error) {
	url := strings.Join([]string{c.config.URL, path}, "/")
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.configureRequest(request)
	return c.client.Do(request)
}

func (c *InstaclustrClient) doPost(path string, body []byte) (*http.Response, error) {
	url := strings.Join([]string{c.config.URL, path}, "/")
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	c.configureRequest(request)
	return c.client.Do(request)
}

func (c *InstaclustrClient) doDelete(path string, body []byte) (*http.Response, error) {
	url := strings.Join([]string{c.config.URL, path}, "/")
	request, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	c.configureRequest(request)
	return c.client.Do(request)
}

func (c *InstaclustrClient) configureRequest(request *http.Request) {
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(c.config.AccessKey, c.config.SecretKey)
}
