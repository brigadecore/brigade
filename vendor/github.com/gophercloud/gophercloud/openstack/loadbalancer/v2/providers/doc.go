/*
Package providers provides information about the supported providers
at OpenStack Octavia Load Balancing service.

Example to List Providers

	allPages, err := providers.List(lbClient).AllPages()
	if err != nil {
		panic(err)
	}

	allProviders, err := providers.ExtractProviders(allPages)
	if err != nil {
		panic(err)
	}

	for _, p := range allProviders {
		fmt.Printf("%+v\n", p)
	}

*/
package providers
