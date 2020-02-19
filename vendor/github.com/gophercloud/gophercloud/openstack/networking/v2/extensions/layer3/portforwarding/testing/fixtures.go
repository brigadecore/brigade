package testing

import "fmt"

const PoFw = `{
      "protocol": "tcp",
      "internal_ip_address": "10.0.0.24",
      "internal_port": 25,
      "internal_port_id": "070ef0b2-0175-4299-be5c-01fea8cca522",
      "external_port": 2229,
      "id": "1798dc82-c0ed-4b79-b12d-4c3c18f90eb2"
    }`

const PoFw_second = `{
      "protocol": "tcp",
      "internal_ip_address": "10.0.0.11",
      "internal_port": 25,
      "internal_port_id": "1238be08-a2a8-4b8d-addf-fb5e2250e480",
      "external_port": 2230,
      "id": "e0a0274e-4d19-4eab-9e12-9e77a8caf3ea"
    }`

var ListResponse = fmt.Sprintf(`
{
    "port_forwardings": [
%s,
%s
    ]
}
`, PoFw, PoFw_second)
