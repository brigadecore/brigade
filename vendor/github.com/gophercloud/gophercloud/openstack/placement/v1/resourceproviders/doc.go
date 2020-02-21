/*
Package resourceproviders creates and lists all resource providers from the OpenStack Placement service.

Example to list resource providers

	allPages, err := resourceproviders.List(placementClient, resourceproviders.ListOpts{}).AllPages()
	if err != nil {
		panic(err)
	}

	allResourceProviders, err := resourceproviders.ExtractResourceProviders(allPages)
	if err != nil {
		panic(err)
	}

	for _, r := range allResourceProviders {
		fmt.Printf("%+v\n", r)
	}

Example to create resource providers

	opts := resourceproviders.CreateOpts{
		Name: "new-rp",
		UUID: "b99b3ab4-3aa6-4fba-b827-69b88b9c544a",
	}

	rp, err := resourceproviders.Create(placementClient, opts).Extract()
	if err != nil {
		panic(err)
	}

*/
package resourceproviders
