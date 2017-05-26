// Package lib contains compiled JavaScript libraries.
//
// See `make generate` for inmformation about generating this library.
package lib

// Script loads a script by path.
//
// The main purpose of this function is to hide the go-bindata library so
// that we can replace it or create a plugable backend in the future.
func Script(name string) (string, error) {
	b, err := Asset(name)
	return string(b), err
}
