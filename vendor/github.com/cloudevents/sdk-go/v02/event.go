package v02

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// Event implements the the CloudEvents specification version 0.2
// https://github.com/cloudevents/spec/blob/v0.2/spec.md
type Event struct {
	// SpecVersion is a mandatory property
	// https://github.com/cloudevents/spec/blob/v0.2/spec.md#cloudeventsversion
	SpecVersion string `json:"specversion" cloudevents:"ce-specversion,required"`
	// EventType is a mandatory property
	// https://github.com/cloudevents/spec/blob/v0.2/spec.md#eventtype
	Type string `json:"type" cloudevents:"ce-type,required"`
	// Source is a mandatory property
	// https://github.com/cloudevents/spec/blob/v0.2/spec.md#source
	Source url.URL `json:"source" cloudevents:"ce-source,required"`
	// ID is a mandatory property
	// https://github.com/cloudevents/spec/blob/v0.2/spec.md#eventid
	ID string `json:"id" cloudevents:"ce-id,required"`
	// Time is an optional property
	// https://github.com/cloudevents/spec/blob/v0.2/spec.md#eventtime
	Time *time.Time `json:"time,omitempty" cloudevents:"ce-time"`
	// SchemaURL is an optional property
	// https://github.com/cloudevents/spec/blob/v0.2/spec.md#schemaurl
	SchemaURL url.URL `json:"schemaurl,omitempty" cloudevents:"ce-schemaurl"`
	// ContentType is an optional property
	// https://github.com/cloudevents/spec/blob/v0.2/spec.md#contenttype
	ContentType string `json:"contenttype,omitempty" cloudevents:"content-type"`
	// Data is an optional property
	// https://github.com/cloudevents/spec/blob/v0.2/spec.md#data-1
	Data interface{} `json:"data,omitempty" cloudevents:",body"`
	// extension an internal map for extension properties not defined in the spec
	extension map[string]interface{}
}

// CloudEventVersion returns the CloudEvents specification version supported by this implementation
func (e Event) CloudEventVersion() (version string) {
	return e.SpecVersion
}

// Get gets a CloudEvent property value
func (e Event) Get(key string) (interface{}, bool) {
	t := reflect.TypeOf(e)
	for i := 0; i < t.NumField(); i++ {
		// Find a matching field by name, ignoring case
		if strings.EqualFold(t.Field(i).Name, key) {
			// return the value of that field
			return reflect.ValueOf(e).Field(i).Interface(), true
		}
	}

	v, ok := e.extension[strings.ToLower(key)]
	return v, ok
}

// GetInt is a convenience method that wraps Get to provide a type checked return value. Ok will be false
// if the property does not exist or the value cannot be converted to an int32.
func (e Event) GetInt(property string) (value int32, ok bool) {
	if val, ok := e.Get(property); ok {
		intVal, ok := val.(int32)
		return intVal, ok
	}

	return 0, false
}

// GetString is a convenience method that wraps Get to provide a type checked return value. Ok will be false
// if the property does not exist or the value cannot be converted to a string.
func (e Event) GetString(property string) (value string, ok bool) {
	if val, ok := e.Get(property); ok {
		stringVal, ok := val.(string)
		return stringVal, ok
	}

	return "", false
}

// GetBinary is a convenience method that wraps Get to provide a type checked return value. Ok will be false
// if the property does not exist or the value cannot be converted to a binary array.
func (e Event) GetBinary(property string) (value []byte, ok bool) {
	if val, ok := e.Get(property); ok {
		binaryArrVal, ok := val.([]byte)
		return binaryArrVal, ok
	}

	return []byte(nil), false
}

// GetMap is a convenience method that wraps Get to provide a type checked return value. Ok will be false
// if the property does not exist or the value cannot be converted to a map.
func (e Event) GetMap(property string) (value map[string]interface{}, ok bool) {
	if val, ok := e.Get(property); ok {
		mapVal, ok := val.(map[string]interface{})
		return mapVal, ok
	}

	return map[string]interface{}(nil), false
}

// GetTime is a convenience method that wraps Get to provide a type checked return value. Ok will be false
// if the property does not exist or the value cannot be converted or parsed into a time.Time.
func (e Event) GetTime(property string) (value *time.Time, ok bool) {
	raw, ok := e.Get(property)
	if !ok {
		return &time.Time{}, false
	}

	switch val := raw.(type) {
	case *time.Time:
		return val, ok
	case time.Time:
		return &val, ok
	case string:
		timestamp, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return &timestamp, false
		}
		return &timestamp, true
	default:
		return &time.Time{}, false
	}
}

// GetURL is a convenience method that wraps Get to provide a type checked return value. Ok will be false
// if the property does not exist or the value cannot be converted or parsed into a url.URL.
func (e Event) GetURL(property string) (value url.URL, ok bool) {
	raw, ok := e.Get(property)
	if !ok {
		return url.URL{}, false
	}

	switch val := raw.(type) {
	case url.URL:
		return val, ok
	case string:
		urlVal, err := url.ParseRequestURI(val)
		if err != nil {
			return url.URL{}, false
		}
		return *urlVal, true
	default:
		return url.URL{}, false
	}
}

// Set sets a CloudEvent property value. If setting a well known field the type of
// value must be assignable to the type of the well known field.
func (e *Event) Set(key string, value interface{}) {
	t := reflect.TypeOf(*e)
	for i := 0; i < t.NumField(); i++ {
		// Find a matching field by name, ignoring case
		if strings.EqualFold(t.Field(i).Name, key) {
			// set that field to the passed in value
			reflect.ValueOf(e).Elem().Field(i).Set(reflect.ValueOf(value))
			return
		}
	}

	// If no matching field, add value to the extension map, creating the map if nil
	if e.extension == nil {
		e.extension = map[string]interface{}{}
	}

	e.extension[strings.ToLower(key)] = value
}

// Properties returns the map of all supported properties in version 0.1.
// The map value says whether particular property is required.
func (e Event) Properties() map[string]bool {
	t := reflect.TypeOf(e)
	props := make(map[string]bool)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		required := false
		if strings.Contains(field.Tag.Get("cloudevents"), "required") {
			required = true
		}

		props[strings.ToLower(field.Name)] = required
	}
	return props
}

type jsonEncodeOpts struct {
	name      string
	omitempty bool
	encoded   bool
	ignored   bool
}

// MarshalJSON marshal an event into json
func (e Event) MarshalJSON() ([]byte, error) {
	output := make(map[string]interface{})

	t := reflect.TypeOf(e)
	eventValue := reflect.ValueOf(e)
	for i := 0; i < t.NumField(); i++ {
		field := eventValue.Field(i)
		if !field.CanInterface() { // if cannot access, i.e. extension
			continue // then ignore it
		}

		jsonOpts := parseJSONTag(t.Field(i))
		// if tag says field is ignored or omitted when empty
		if jsonOpts.ignored ||
			(jsonOpts.omitempty && isZero(field, t.Field(i).Type)) {
			continue // then ignore it
		}

		val := field.Interface()
		if url, ok := val.(url.URL); ok {
			val = url.String()
		}

		output[jsonOpts.name] = val
	}

	for k, v := range e.extension {
		output[k] = v
	}

	return json.Marshal(output)
}

// UnmarshalJSON override default json unmarshall
func (e *Event) UnmarshalJSON(b []byte) error {
	// Unmarshal the raw data into an intermediate map
	var intermediate map[string]interface{}
	if err := json.Unmarshal(b, &intermediate); err != nil {
		return err
	}

	t := reflect.TypeOf(*e)
	target := reflect.New(t).Elem()
	for i := 0; i < t.NumField(); i++ {
		targetField := target.Field(i) // allows us to modify the target value
		if !targetField.CanSet() {     // if cannot be set, i.e. extension
			continue // ignore it
		}

		structField := t.Field(i) // contains type info
		jsonOpts := parseJSONTag(structField)
		if jsonOpts.ignored {
			continue
		}

		// if prop exists in map convert it and set to the target's field.
		if mapVal, ok := intermediate[jsonOpts.name]; ok {
			assignToFieldByType(&targetField, mapVal)
		}

		// remove processed field
		delete(intermediate, jsonOpts.name)
	}

	*e = target.Interface().(Event)

	if len(intermediate) == 0 {
		return nil
	}

	// add any left over fields to the extensions map
	e.extension = make(map[string]interface{}, len(intermediate))
	for k, v := range intermediate {
		e.extension[k] = v
	}

	return nil
}

func isZero(fieldValue reflect.Value, fieldType reflect.Type) bool {
	return reflect.DeepEqual(fieldValue.Interface(), reflect.Zero(fieldType).Interface())
}

// assignToFieldByType assigns the value to the given field Value converting to the correct type if necessary
func assignToFieldByType(field *reflect.Value, value interface{}) error {
	var val reflect.Value
	switch field.Type().Kind() {
	case reflect.TypeOf((*time.Time)(nil)).Kind():
		timestamp, err := time.Parse(time.RFC3339, value.(string))
		if err != nil {
			return err
		}
		val = reflect.ValueOf(&timestamp)
	case reflect.TypeOf((*url.URL)(nil)).Elem().Kind():
		url, err := url.ParseRequestURI(value.(string))
		if err != nil {
			return err
		}
		val = reflect.ValueOf(*url)
	default:
		val = reflect.ValueOf(value)
	}

	field.Set(val)
	return nil
}

func parseJSONTag(field reflect.StructField) jsonEncodeOpts {
	tag := field.Tag.Get("json")
	opts := jsonEncodeOpts{}
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

	if strings.Contains(options[1], "omitempty") {
		opts.omitempty = true
	}

	if strings.Contains(options[1], "string") {
		opts.encoded = true
	}

	return opts
}
