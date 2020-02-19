package testing

import (
	"testing"
	"time"

	"github.com/gophercloud/gophercloud/openstack/blockstorage/apiversions"
	th "github.com/gophercloud/gophercloud/testhelper"
	"github.com/gophercloud/gophercloud/testhelper/client"
)

func TestListVersions(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockListResponse(t)

	allVersions, err := apiversions.List(client.ServiceClient()).AllPages()
	th.AssertNoErr(t, err)
	actual, err := apiversions.ExtractAPIVersions(allVersions)
	th.AssertNoErr(t, err)

	expected := []apiversions.APIVersion{
		{
			ID:      "v1.0",
			Status:  "DEPRECATED",
			Updated: time.Date(2016, 5, 2, 20, 25, 19, 0, time.UTC),
		},
		{
			ID:      "v2.0",
			Status:  "SUPPORTED",
			Updated: time.Date(2014, 6, 28, 12, 20, 21, 0, time.UTC),
		},
		{
			ID:         "v3.0",
			MinVersion: "3.0",
			Status:     "CURRENT",
			Updated:    time.Date(2016, 2, 8, 12, 20, 21, 0, time.UTC),
			Version:    "3.27",
		},
	}

	th.AssertDeepEquals(t, expected, actual)
}

func TestListOldVersions(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockListOldResponse(t)

	allVersions, err := apiversions.List(client.ServiceClient()).AllPages()
	th.AssertNoErr(t, err)
	actual, err := apiversions.ExtractAPIVersions(allVersions)
	th.AssertNoErr(t, err)

	expected := []apiversions.APIVersion{
		{
			ID:      "v1.0",
			Status:  "CURRENT",
			Updated: time.Date(2012, 1, 4, 11, 33, 21, 0, time.UTC),
		},
		{
			ID:      "v2.0",
			Status:  "CURRENT",
			Updated: time.Date(2012, 11, 21, 11, 33, 21, 0, time.UTC),
		},
	}

	th.AssertDeepEquals(t, expected, actual)
}

func TestGetVersion(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockListResponse(t)

	allVersions, err := apiversions.List(client.ServiceClient()).AllPages()
	th.AssertNoErr(t, err)
	actual, err := apiversions.ExtractAPIVersion(allVersions, "v3.0")
	th.AssertNoErr(t, err)

	expected := apiversions.APIVersion{
		ID:         "v3.0",
		MinVersion: "3.0",
		Status:     "CURRENT",
		Updated:    time.Date(2016, 2, 8, 12, 20, 21, 0, time.UTC),
		Version:    "3.27",
	}

	th.AssertEquals(t, actual.ID, expected.ID)
	th.AssertEquals(t, actual.Status, expected.Status)
	th.AssertEquals(t, actual.Updated, expected.Updated)
}

func TestGetOldVersion(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	MockListOldResponse(t)

	allVersions, err := apiversions.List(client.ServiceClient()).AllPages()
	th.AssertNoErr(t, err)
	actual, err := apiversions.ExtractAPIVersion(allVersions, "v2.0")
	th.AssertNoErr(t, err)

	expected := apiversions.APIVersion{
		ID:         "v2.0",
		MinVersion: "",
		Status:     "CURRENT",
		Updated:    time.Date(2012, 11, 21, 11, 33, 21, 0, time.UTC),
		Version:    "",
	}

	th.AssertEquals(t, actual.ID, expected.ID)
	th.AssertEquals(t, actual.Status, expected.Status)
	th.AssertEquals(t, actual.Updated, expected.Updated)
}
