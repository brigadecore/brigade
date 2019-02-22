package webhook

import (
	"fmt"
	"log"

	"github.com/Azure/brigade/pkg/brigade"
)

// validateGenericGatewaySecret will return an error if given Project does not have a GenericGatewaySecret or if the provided secret is wrong
// Otherwise, it will simply return nil
func validateGenericGatewaySecret(proj *brigade.Project, secret string) error {
	// if the secret is "" (probably i) due to a Brigade upgrade or ii) user did not create a Generic Gateway secret during `brig project create`)
	// refuse to serve it, so Brigade admin will be forced to update the project with a non-empty secret
	if proj.GenericGatewaySecret == "" {
		log.Printf("Secret for project %s is empty, please update it and try again", proj.ID)
		return fmt.Errorf("secret for this Brigade Project is empty, refusing to serve, please inform your Brigade admin")
	}

	// compare secrets
	if secret != proj.GenericGatewaySecret {
		log.Printf("Secret %s for project %s is wrong", secret, proj.ID)
		return fmt.Errorf("secret is wrong")
	}

	return nil
}
