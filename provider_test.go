package main

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/builtin/providers/aws"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]terraform.ResourceProvider{
		"instaclustr": testAccProvider,
		"aws":         aws.Provider(),
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("INSTACLUSTR_ACCESS_KEY"); v == "" {
		t.Fatal("INSTACLUSTR_ACCESS_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("INSTACLUSTR_SECRET_KEY"); v == "" {
		t.Fatal("INSTACLUSTR_SECRET_KEY must be set for acceptance tests")
	}
}
