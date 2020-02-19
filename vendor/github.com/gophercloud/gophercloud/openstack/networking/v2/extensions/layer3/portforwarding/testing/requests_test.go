package testing

import (
	"fmt"
	"net/http"
	"testing"

	fake "github.com/gophercloud/gophercloud/openstack/networking/v2/common"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/portforwarding"
	"github.com/gophercloud/gophercloud/pagination"
	th "github.com/gophercloud/gophercloud/testhelper"
)

func TestPortForwardingList(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v2.0/floatingips/2f95fd2b-9f6a-4e8e-9e9a-2cbe286cbf9e/port_forwardings", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, ListResponse)
	})

	count := 0

	portforwarding.List(fake.ServiceClient(), portforwarding.ListOpts{}, "2f95fd2b-9f6a-4e8e-9e9a-2cbe286cbf9e").EachPage(func(page pagination.Page) (bool, error) {
		count++
		actual, err := portforwarding.ExtractPortForwardings(page)
		if err != nil {
			t.Errorf("Failed to extract port forwardings: %v", err)
			return false, err
		}

		expected := []portforwarding.PortForwarding{
			{
				Protocol:          "tcp",
				InternalIPAddress: "10.0.0.24",
				InternalPort:      25,
				InternalPortID:    "070ef0b2-0175-4299-be5c-01fea8cca522",
				ExternalPort:      2229,
				ID:                "1798dc82-c0ed-4b79-b12d-4c3c18f90eb2",
			},
			{
				Protocol:          "tcp",
				InternalIPAddress: "10.0.0.11",
				InternalPort:      25,
				InternalPortID:    "1238be08-a2a8-4b8d-addf-fb5e2250e480",
				ExternalPort:      2230,
				ID:                "e0a0274e-4d19-4eab-9e12-9e77a8caf3ea",
			},
		}

		th.CheckDeepEquals(t, expected, actual)

		return true, nil
	})

	if count != 1 {
		t.Errorf("Expected 1 page, got %d", count)
	}
}

func TestCreate(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v2.0/floatingips/2f95fd2b-9f6a-4e8e-9e9a-2cbe286cbf9e/port_forwardings", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "POST")
		th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)
		th.TestHeader(t, r, "Content-Type", "application/json")
		th.TestHeader(t, r, "Accept", "application/json")
		th.TestJSONRequest(t, r, `
{
  	"port_forwarding": {
      "protocol": "tcp",
      "internal_ip_address": "10.0.0.11",
      "internal_port": 25,
      "internal_port_id": "1238be08-a2a8-4b8d-addf-fb5e2250e480",
      "external_port": 2230
  }
}
		
			`)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		fmt.Fprintf(w, `
{
	"port_forwarding": {
    		"protocol": "tcp",
    		"internal_ip_address": "10.0.0.11",
    		"internal_port": 25,
    		"internal_port_id": "1238be08-a2a8-4b8d-addf-fb5e2250e480",
    		"external_port": 2230,
    		"id": "725ade3c-9760-4880-8080-8fc2dbab9acc"
  }
}`)
	})

	options := portforwarding.CreateOpts{
		Protocol:          "tcp",
		InternalIPAddress: "10.0.0.11",
		InternalPort:      25,
		ExternalPort:      2230,
		InternalPortID:    "1238be08-a2a8-4b8d-addf-fb5e2250e480",
	}

	pf, err := portforwarding.Create(fake.ServiceClient(), "2f95fd2b-9f6a-4e8e-9e9a-2cbe286cbf9e", options).Extract()
	th.AssertNoErr(t, err)

	th.AssertEquals(t, "725ade3c-9760-4880-8080-8fc2dbab9acc", pf.ID)
	th.AssertEquals(t, "10.0.0.11", pf.InternalIPAddress)
	th.AssertEquals(t, 25, pf.InternalPort)
	th.AssertEquals(t, "1238be08-a2a8-4b8d-addf-fb5e2250e480", pf.InternalPortID)
	th.AssertEquals(t, 2230, pf.ExternalPort)
	th.AssertEquals(t, "tcp", pf.Protocol)
}

func TestGet(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v2.0/floatingips/2f245a7b-796b-4f26-9cf9-9e82d248fda7/port_forwardings/725ade3c-9760-4880-8080-8fc2dbab9acc", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, `
{
  "port_forwarding": {
    "protocol": "tcp",
    "internal_ip_address": "10.0.0.11",
    "internal_port": 25,
    "internal_port_id": "1238be08-a2a8-4b8d-addf-fb5e2250e480",
    "external_port": 2230,
    "id": "725ade3c-9760-4880-8080-8fc2dbab9acc"
  }
}
      `)
	})

	pf, err := portforwarding.Get(fake.ServiceClient(), "2f245a7b-796b-4f26-9cf9-9e82d248fda7", "725ade3c-9760-4880-8080-8fc2dbab9acc").Extract()
	th.AssertNoErr(t, err)

	th.AssertEquals(t, "tcp", pf.Protocol)
	th.AssertEquals(t, "725ade3c-9760-4880-8080-8fc2dbab9acc", pf.ID)
	th.AssertEquals(t, "10.0.0.11", pf.InternalIPAddress)
	th.AssertEquals(t, 25, pf.InternalPort)
	th.AssertEquals(t, "1238be08-a2a8-4b8d-addf-fb5e2250e480", pf.InternalPortID)
	th.AssertEquals(t, 2230, pf.ExternalPort)
}

func TestDelete(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v2.0/floatingips/2f245a7b-796b-4f26-9cf9-9e82d248fda7/port_forwardings/725ade3c-9760-4880-8080-8fc2dbab9acc", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "DELETE")
		th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)
		w.WriteHeader(http.StatusNoContent)
	})

	res := portforwarding.Delete(fake.ServiceClient(), "2f245a7b-796b-4f26-9cf9-9e82d248fda7", "725ade3c-9760-4880-8080-8fc2dbab9acc")
	th.AssertNoErr(t, res.Err)
}

func TestUpdate(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/v2.0/floatingips/2f245a7b-796b-4f26-9cf9-9e82d248fda7/port_forwardings/725ade3c-9760-4880-8080-8fc2dbab9acc", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "PUT")
		th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)
		th.TestHeader(t, r, "Content-Type", "application/json")
		th.TestHeader(t, r, "Accept", "application/json")
		th.TestJSONRequest(t, r, `
{
  "port_forwarding": {
    "protocol": "udp",
    "internal_port": 37,
    "internal_port_id": "99889dc2-19a7-4edb-b9d0-d2ace8d1e144",
    "external_port": 1960
  }
}
			`)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, `
{
  "port_forwarding": {
    "protocol": "udp",
    "internal_ip_address": "10.0.0.14",
    "internal_port": 37,
    "internal_port_id": "99889dc2-19a7-4edb-b9d0-d2ace8d1e144",
    "external_port": 1960,
    "id": "725ade3c-9760-4880-8080-8fc2dbab9acc"
  }
}
`)
	})

	updatedProtocol := "udp"
	updatedInternalPort := 37
	updatedInternalPortID := "99889dc2-19a7-4edb-b9d0-d2ace8d1e144"
	updatedExternalPort := 1960
	options := portforwarding.UpdateOpts{
		Protocol:       updatedProtocol,
		InternalPort:   updatedInternalPort,
		InternalPortID: updatedInternalPortID,
		ExternalPort:   updatedExternalPort,
	}

	actual, err := portforwarding.Update(fake.ServiceClient(), "2f245a7b-796b-4f26-9cf9-9e82d248fda7", "725ade3c-9760-4880-8080-8fc2dbab9acc", options).Extract()
	th.AssertNoErr(t, err)
	expected := portforwarding.PortForwarding{
		Protocol:          "udp",
		InternalIPAddress: "10.0.0.14",
		InternalPort:      37,
		ID:                "725ade3c-9760-4880-8080-8fc2dbab9acc",
		InternalPortID:    "99889dc2-19a7-4edb-b9d0-d2ace8d1e144",
		ExternalPort:      1960,
	}
	th.AssertDeepEquals(t, expected, *actual)
}
