package shelveunshelve

import "github.com/gophercloud/gophercloud"

// ShelveResult is the response from a Shelve operation. Call its ExtractErr
// method to determine if the request succeeded or failed.
type ShelveResult struct {
	gophercloud.ErrResult
}

// ShelveOffloadResult is the response from a Shelve operation. Call its ExtractErr
// method to determine if the request succeeded or failed.
type ShelveOffloadResult struct {
	gophercloud.ErrResult
}

// UnshelveResult is the response from Stop operation. Call its ExtractErr
// method to determine if the request succeeded or failed.
type UnshelveResult struct {
	gophercloud.ErrResult
}
