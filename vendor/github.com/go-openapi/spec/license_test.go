// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spec

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var license = License{
	LicenseProps:     LicenseProps{Name: "the name", URL: "the url"},
	VendorExtensible: VendorExtensible{Extensions: map[string]interface{}{"x-license": "custom term"}}}

const licenseJSON = `{
	"name": "the name",
	"url": "the url",
	"x-license": "custom term"
}`

func TestIntegrationLicense(t *testing.T) {

	const licenseYAML = "name: the name\nurl: the url\n"

	b, err := json.MarshalIndent(license, "", "\t")
	if assert.NoError(t, err) {
		assert.Equal(t, licenseJSON, string(b))
	}

	actual := License{}
	err = json.Unmarshal([]byte(licenseJSON), &actual)
	if assert.NoError(t, err) {
		assert.EqualValues(t, license, actual)
	}
}
