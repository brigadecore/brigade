package testing

import (
	"fmt"
	"net/http"
	"testing"

	th "github.com/gophercloud/gophercloud/testhelper"
	"github.com/gophercloud/gophercloud/testhelper/client"
)

const APIListResponse = `
{
    "versions": [
        {
            "id": "v1.0",
            "links": [
                {
                    "href": "http://docs.openstack.org/",
                    "rel": "describedby",
                    "type": "text/html"
                },
                {
                    "href": "http://localhost:8776/v1/",
                    "rel": "self"
                }
            ],
            "media-types": [
                {
                    "base": "application/json",
                    "type": "application/vnd.openstack.volume+json;version=1"
                }
            ],
            "min_version": "",
            "status": "DEPRECATED",
            "updated": "2016-05-02T20:25:19Z",
            "version": ""
        },
        {
            "id": "v2.0",
            "links": [
                {
                    "href": "http://docs.openstack.org/",
                    "rel": "describedby",
                    "type": "text/html"
                },
                {
                    "href": "http://localhost:8776/v2/",
                    "rel": "self"
                }
            ],
            "media-types": [
                {
                    "base": "application/json",
                    "type": "application/vnd.openstack.volume+json;version=1"
                }
            ],
            "min_version": "",
            "status": "SUPPORTED",
            "updated": "2014-06-28T12:20:21Z",
            "version": ""
        },
        {
            "id": "v3.0",
            "links": [
                {
                    "href": "http://docs.openstack.org/",
                    "rel": "describedby",
                    "type": "text/html"
                },
                {
                    "href": "http://localhost:8776/v3/",
                    "rel": "self"
                }
            ],
            "media-types": [
                {
                    "base": "application/json",
                    "type": "application/vnd.openstack.volume+json;version=1"
                }
            ],
            "min_version": "3.0",
            "status": "CURRENT",
            "updated": "2016-02-08T12:20:21Z",
            "version": "3.27"
        }
    ]
}
`

const APIListOldResponse = `
{
  "versions": [
    {
      "status": "CURRENT",
      "updated": "2012-01-04T11:33:21Z",
      "id": "v1.0",
      "links": [
        {
          "href": "http://23.253.228.211:8776/v1/",
          "rel": "self"
        }
      ]
      },
    {
      "status": "CURRENT",
      "updated": "2012-11-21T11:33:21Z",
      "id": "v2.0",
      "links": [
        {
          "href": "http://23.253.228.211:8776/v2/",
          "rel": "self"
        }
      ]
    }
  ]
}`

func MockListResponse(t *testing.T) {
	th.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, APIListResponse)
	})
}

func MockListOldResponse(t *testing.T) {
	th.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, APIListOldResponse)
	})
}
