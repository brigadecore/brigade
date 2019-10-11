package mocks

import (
	"github.com/brigadecore/brigade/pkg/storage"
)

// Store is the wrapper for brigade store interface for mocks.
type Store interface {
	storage.Store
}
