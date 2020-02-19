/*
Package shelveunshelve provides functionality to start and stop servers that have
been provisioned by the OpenStack Compute service.

Example to Shelve, Shelve-offload and Unshelve a Server

	serverID := "47b6b7b7-568d-40e4-868c-d5c41735532e"

	err := shelveunshelve.Shelve(computeClient, serverID).ExtractErr()
	if err != nil {
		panic(err)
	}

	err := shelveunshelve.ShelveOffload(computeClient, serverID).ExtractErr()
	if err != nil {
		panic(err)
	}

	err := shelveunshelve.Unshelve(computeClient, serverID, nil).ExtractErr()
	if err != nil {
		panic(err)
	}
*/
package shelveunshelve
