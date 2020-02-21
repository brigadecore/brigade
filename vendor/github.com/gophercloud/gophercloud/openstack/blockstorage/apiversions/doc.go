/*
Package apiversions provides information and interaction with the different
API versions for the OpenStack Block Storage service, code-named Cinder.

Example of Retrieving all API Versions

	allPages, err := apiversions.List(client).AllPages()
	if err != nil {
		panic("Unable to get API versions: %s", err)
	}

	allVersions, err := apiversions.ExtractAPIVersions(allPages)
	if err != nil {
		panic("Unable to extract API versions: %s", err)
	}

	for _, version := range versions {
		fmt.Printf("%+v\n", version)
	}


Example of Retrieving an API Version

	version, err := apiversions.Get(client, "v3").Extract()
	if err != nil {
		panic("Unable to get API version: %s", err)
	}

	fmt.Printf("%+v\n", version)
*/
package apiversions
