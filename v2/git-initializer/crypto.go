package main

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"

	"github.com/pkg/errors"
)

// decryptKey decrypts a PEM-encoded RSA private key and returns a PEM-encoded
// RSA private key sans encryption. This is useful for cases where the key is
// being used in conjunction with the git CLI and we wish to avoid interactively
// requesting the key's passphrase.
func decryptKey(key, pass string) (string, error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return "", errors.Errorf("key is not PEM-encoded")
	}
	const supportedType = "RSA PRIVATE KEY"
	if block.Type != supportedType {
		return "", errors.Errorf(
			"unsupported key type %q; only %q is supported",
			block.Type,
			supportedType,
		)
	}
	var err error
	if block.Bytes, err = x509.DecryptPEMBlock(block, []byte(pass)); err != nil {
		return "", errors.Wrap(err, "error decrypting private key")
	}
	block.Headers = nil
	buf := &bytes.Buffer{}
	if err = pem.Encode(buf, block); err != nil {
		return "", errors.Wrap(err, "error PEM-encoding decrypted private key")
	}
	return buf.String(), nil
}
