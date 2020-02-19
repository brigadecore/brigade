package apiversions

import (
	"net/url"

	"github.com/gophercloud/gophercloud"
)

func listURL(c *gophercloud.ServiceClient) string {
	u, _ := url.Parse(c.ServiceURL(""))
	u.Path = "/"
	return u.String()
}
