package v02

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"reflect"

	"github.com/cloudevents/sdk-go"
)

// HTTPMarshaller A struct representing the v02 version of the HTTPMarshaller
type HTTPMarshaller struct {
	converters []cloudevents.HTTPCloudEventConverter
}

// NewDefaultHTTPMarshaller creates a new v02 HTTPMarshaller prepopulated with the Binary and JSON
// CloudEvent converters
func NewDefaultHTTPMarshaller() cloudevents.HTTPMarshaller {
	return NewHTTPMarshaller(
		NewJSONHTTPCloudEventConverter(),
		NewBinaryHTTPCloudEventConverter())
}

// NewHTTPMarshaller creates a new HTTPMarshaller with the given HTTPCloudEventConverters
func NewHTTPMarshaller(converters ...cloudevents.HTTPCloudEventConverter) cloudevents.HTTPMarshaller {
	return &HTTPMarshaller{
		converters: converters,
	}
}

// FromRequest creates a new CloudEvent from an http Request
func (e HTTPMarshaller) FromRequest(req *http.Request) (cloudevents.Event, error) {
	if req == nil {
		return nil, cloudevents.IllegalArgumentError("req")
	}

	mimeType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("error parsing request content type: %s", err.Error())
	}

	for _, v := range e.converters {
		if v.CanRead(reflect.TypeOf(Event{}), mimeType) {
			return v.Read(reflect.TypeOf(Event{}), req)
		}
	}
	return nil, cloudevents.ContentTypeNotSupportedError(mimeType)
}

// ToRequest populates an http Request with the given CloudEvent
func (e HTTPMarshaller) ToRequest(req *http.Request, event cloudevents.Event) error {
	if req == nil {
		return cloudevents.IllegalArgumentError("req")
	}

	if event == nil {
		return cloudevents.IllegalArgumentError("event")
	}

	v02Event := event.(*Event)

	contentType := v02Event.ContentType
	if contentType == "" {
		contentType = "application/cloudevents+json"
	}

	mimeType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("error parsing event content type: %s", err.Error())
	}

	for _, v := range e.converters {
		if v.CanWrite(reflect.TypeOf(Event{}), mimeType) {
			err := v.Write(reflect.TypeOf(Event{}), req, event)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return cloudevents.ContentTypeNotSupportedError(mimeType)
}

// jsonhttpCloudEventConverter new converter for reading/writing CloudEvents to JSON
type jsonhttpCloudEventConverter struct {
	supportedMediaTypes      map[string]bool
	supportedMediaTypesSlice []string
}

// NewJSONHTTPCloudEventConverter creates a new JSONHTTPCloudEventConverter
func NewJSONHTTPCloudEventConverter() cloudevents.HTTPCloudEventConverter {
	mediaTypes := map[string]bool{
		"application/cloudevents+json": true,
	}

	return &jsonhttpCloudEventConverter{
		supportedMediaTypes: mediaTypes,
	}
}

// CanRead specifies if this converter can read the given mediaType into a given reflect.Type
func (j *jsonhttpCloudEventConverter) CanRead(t reflect.Type, mediaType string) bool {
	return reflect.TypeOf(Event{}) == t && j.supportedMediaTypes[mediaType]
}

// CanWrite specifies if this converter can write the given Type into the given mediaType
func (j *jsonhttpCloudEventConverter) CanWrite(t reflect.Type, mediaType string) bool {
	return reflect.TypeOf(Event{}) == t && j.supportedMediaTypes[mediaType]
}

func (j *jsonhttpCloudEventConverter) Read(t reflect.Type, req *http.Request) (cloudevents.Event, error) {
	e := reflect.New(t).Interface()
	err := json.NewDecoder(req.Body).Decode(e)

	if err != nil {
		return nil, fmt.Errorf("error parsing request: %s", err.Error())
	}

	ret := e.(*Event)
	return ret, nil
}

func (j *jsonhttpCloudEventConverter) Write(t reflect.Type, req *http.Request, event cloudevents.Event) error {
	buffer := bytes.Buffer{}
	if err := json.NewEncoder(&buffer).Encode(event); err != nil {
		return err
	}

	req.Body = ioutil.NopCloser(&buffer)
	req.ContentLength = int64(buffer.Len())
	req.GetBody = func() (io.ReadCloser, error) {
		reader := bytes.NewReader(buffer.Bytes())
		return ioutil.NopCloser(reader), nil
	}

	req.Header.Set("Content-Type", "application/cloudevents+json")
	return nil
}

// BinaryHTTPCloudEventConverter a converter for reading/writing CloudEvents into the binary format
type binaryHTTPCloudEventConverter struct {
	supportedMediaTypes map[string]bool
}

// NewBinaryHTTPCloudEventConverter creates a new BinaryHTTPCloudEventConverter
func NewBinaryHTTPCloudEventConverter() cloudevents.HTTPCloudEventConverter {
	mediaTypes := map[string]bool{
		"application/json":         true,
		"application/xml":          true,
		"application/octet-stream": true,
	}

	return &binaryHTTPCloudEventConverter{
		supportedMediaTypes: mediaTypes,
	}
}

// CanRead specifies if this converter can read the given mediaType into a given reflect.Type
func (b *binaryHTTPCloudEventConverter) CanRead(t reflect.Type, mediaType string) bool {
	return reflect.TypeOf(Event{}) == t && b.supportedMediaTypes[mediaType]
}

// CanWrite specifies if this converter can write the given Type into the given mediaType
func (b *binaryHTTPCloudEventConverter) CanWrite(t reflect.Type, mediaType string) bool {
	return reflect.TypeOf(Event{}) == t && b.supportedMediaTypes[mediaType]
}

func (b *binaryHTTPCloudEventConverter) Read(t reflect.Type, req *http.Request) (cloudevents.Event, error) {
	var event Event
	if err := event.UnmarshalBinary(req); err != nil {
		return nil, err
	}

	return &event, nil
}

func (b *binaryHTTPCloudEventConverter) Write(t reflect.Type, req *http.Request, event cloudevents.Event) error {
	e := event.(*Event)
	if err := e.MarshalBinary(req); err != nil {
		return err
	}
	return nil
}
