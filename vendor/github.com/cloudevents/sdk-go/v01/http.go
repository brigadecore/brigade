package v01

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

// HTTPFormat type wraps supported modes of formatting CloudEvent as HTTP request.
// Currently, only binary mode and structured mode with JSON encoding are supported.
type HTTPFormat string

const (
	// FormatBinary corresponds to Binary mode in CloudEvents HTTP transport binding.
	// https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/http-transport-binding.md#31-binary-content-mode
	FormatBinary HTTPFormat = "binary"
	// FormatJSON corresponds to Structured mode using JSON encoding.
	// https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/http-transport-binding.md#32-structured-content-mode
	FormatJSON HTTPFormat = "json"
)

const (
	ceContentType = "application/cloudevents"
)

// MarshalBinary marshal an event in the binary representation
func (e Event) MarshalBinary(req *http.Request) error {
	t := reflect.TypeOf(e)
	eventValue := reflect.ValueOf(e)

	header := http.Header{}

	for i := 0; i < t.NumField(); i++ {
		field := eventValue.Field(i)
		structField := t.Field(i)
		if !field.CanInterface() || isZero(field, structField.Type) { //ignore unexported/unset fields
			continue
		}

		opts := parseCloudEventsTag(structField)

		if opts.body { // ignore the field designated as the body
			continue
		}

		val := field.Interface()
		if url, ok := val.(url.URL); ok {
			val = url.String()
		}
		header.Set(opts.name, fmt.Sprintf("%v", val))
	}

	for k, v := range e.extension {
		header.Set(headerize(k), fmt.Sprintf("%v", v))
	}

	req.Header = header

	data := e.Data
	buffer := &bytes.Buffer{}
	switch req.Header.Get("Content-Type") {
	case "application/json":
		json.NewEncoder(buffer).Encode(data)
	case "application/xml":
		xml.NewEncoder(buffer).Encode(data)
	default:
		buffer = bytes.NewBuffer(data.([]byte))
	}

	req.Body = ioutil.NopCloser(buffer)
	req.ContentLength = int64(buffer.Len())
	req.GetBody = func() (io.ReadCloser, error) {
		reader := bytes.NewReader(buffer.Bytes())
		return ioutil.NopCloser(reader), nil
	}

	return nil
}

// UnmarshalBinary unmarshal an event from the binary representation
func (e *Event) UnmarshalBinary(req *http.Request) error {
	t := reflect.TypeOf(*e)
	target := reflect.New(t).Elem()

	extension := make(map[string]interface{})

	var fieldNames = make(map[string]string, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		opts := parseCloudEventsTag(t.Field(i))
		if opts.body {
			continue
		}
		fieldNames[opts.name] = t.Field(i).Name
	}

	for k, v := range req.Header {
		lowerHeader := strings.ToLower(k)
		fieldName, ok := fieldNames[lowerHeader]

		// if field is unknown and prepended with "ce-" add this header to our extensions bag
		if !ok {
			if strings.HasPrefix(lowerHeader, "ce-") {
				extension[strings.TrimPrefix(lowerHeader, "ce-")] = v[0]
			}
			continue
		}

		field := target.FieldByName(fieldName)
		assignToFieldByType(&field, v[0])
	}

	if req.ContentLength != 0 {
		var data interface{}
		var err error
		switch req.Header.Get("Content-Type") {
		case "application/json":
			err = json.NewDecoder(req.Body).Decode(&data)
		case "application/xml":
			err = xml.NewDecoder(req.Body).Decode(&data)
		default:
			buffer := &bytes.Buffer{}
			buffer.ReadFrom(req.Body)
			data = buffer.Bytes()
		}

		if err != nil {
			return fmt.Errorf("error reading request body: %s", err.Error())
		}

		target.FieldByName("Data").Set(reflect.ValueOf(data))
	}

	targetEvent := target.Interface().(Event)
	targetEvent.extension = extension
	*e = targetEvent

	return nil
}

type cloudeventsEncodeOpts struct {
	name     string
	required bool
	body     bool
	ignored  bool
}

func parseCloudEventsTag(field reflect.StructField) cloudeventsEncodeOpts {
	tag := field.Tag.Get("cloudevents")
	opts := cloudeventsEncodeOpts{}
	if tag == "-" {
		opts.ignored = true
		return opts
	}

	options := strings.SplitN(tag, ",", 2)

	opts.name = options[0]
	if opts.name == "" {
		opts.name = field.Name
	}

	if len(options) == 1 {
		return opts
	}

	if strings.Contains(options[1], "required") {
		opts.required = true
	}

	if strings.Contains(options[1], "body") {
		opts.body = true
	}

	return opts
}

func headerize(property string) string {
	return "ce-" + strings.Title(property)
}
