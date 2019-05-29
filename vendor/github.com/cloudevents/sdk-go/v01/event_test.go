package v01_test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/v01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEvent(t *testing.T) {
	timestamp, err := time.Parse(time.RFC3339, "2018-04-05T17:31:00Z")
	require.NoError(t, err)

	event := &v01.Event{
		EventType: "com.example.someevent",
		Source: url.URL{
			Path: "/mycontext",
		},
		EventID:   "1234-1234-1234",
		EventTime: &timestamp,
		SchemaURL: url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/schema",
		},
		ContentType: "application/json",
		Data:        map[string]interface{}{"key": "value"},
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Errorf("JSON Error received: %v", err)
	}
	fmt.Printf("%s", data)

	eventUnmarshaled := &v01.Event{}
	json.Unmarshal(data, eventUnmarshaled)
	assert.EqualValues(t, event, eventUnmarshaled)
}

func TestGetSet(t *testing.T) {
	event := &v01.Event{
		EventType: "com.example.someevent",
		Source: url.URL{
			Path: "/mycontext",
		},
		EventID:   "1234-1234-1234",
		EventTime: nil,
		SchemaURL: url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/schema",
		},
		ContentType: "application/json",
		Data:        map[string]interface{}{"key": "value"},
	}

	value, ok := event.Get("nonexistent")
	assert.False(t, ok, "ok should be false for nonexistent key, but isn't")
	assert.Nil(t, value, "value for nonexistent key should be nil, but isn't")

	value, ok = event.Get("contentType")
	assert.True(t, ok, "ok for existing key should be true, but isn't")
	assert.Equal(t, "application/json", value, "value for contentType should be application/json, but is %s", value)

	event.Set("eventType", "newType")
	assert.Equal(t, "newType", event.EventType, "expected eventType to be 'newType', got %s", event.EventType)

	event.Set("ext", "somevalue")
	value, ok = event.Get("ext")
	assert.True(t, ok, "ok for ext key should be true, but isn't")
	assert.Equal(t, "somevalue", value, "value for ext key should be 'somevalue', but is %s", value)
}

func TestProperties(t *testing.T) {
	event := v01.Event{}

	props := event.Properties()

	assert.True(t, props["eventid"])
	delete(props, "eventid")
	assert.True(t, props["source"])
	delete(props, "source")
	assert.True(t, props["eventtype"])
	delete(props, "eventtype")
	assert.True(t, props["cloudeventsversion"])
	delete(props, "cloudeventsversion")

	for k, v := range props {
		assert.False(t, v, "property %s should not be required.", k)
	}
}

func TestGetIntOk(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}
	var expected int32 = 100
	principal.Set("myint", expected)

	actual, ok := principal.GetInt("myint")

	assert.True(t, ok)
	assert.Equal(t, expected, actual)
}

func TestGetIntWrongType(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}
	principal.Set("notint", "not an int")

	actual, ok := principal.GetInt("notint")

	assert.False(t, ok)
	assert.Equal(t, int32(0), actual)
}

func TestGetIntMissing(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}

	actual, ok := principal.GetInt("missing")

	assert.False(t, ok)
	assert.Equal(t, int32(0), actual)
}

func TestGetStringOk(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{
		EventType: "com.example.someevent",
	}

	actual, ok := principal.GetString("EventType")

	assert.True(t, ok)
	assert.Equal(t, actual, "com.example.someevent")
}

func TestGetStringWrongType(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}

	principal.Set("mystringfail", 100)

	actual, ok := principal.GetString("mystringfail")

	assert.False(t, ok)
	assert.Equal(t, "", actual)
}

func TestGetTimeOk(t *testing.T) {
	expected := time.Now()
	var principal cloudevents.Event = &v01.Event{
		EventTime: &expected,
	}

	actual, ok := principal.GetTime("EventTime")

	assert.True(t, ok)
	assert.Equal(t, &expected, actual)
}

func TestGetTimeExtensionOk(t *testing.T) {
	expected := time.Now()
	var principal cloudevents.Event = &v01.Event{}
	principal.Set("mytime", expected)

	actual, ok := principal.GetTime("mytime")

	assert.True(t, ok)
	assert.Equal(t, expected, *actual)
}

func TestGetTimePointerExtensionOk(t *testing.T) {
	expected := time.Now()
	var principal cloudevents.Event = &v01.Event{}
	principal.Set("mytime", &expected)

	actual, ok := principal.GetTime("mytime")

	assert.True(t, ok)
	assert.Equal(t, &expected, actual)
}

func TestGetTimeStringOk(t *testing.T) {
	expected := time.Now().Format(time.RFC3339)
	var principal cloudevents.Event = &v01.Event{}
	principal.Set("mytimestring", expected)

	actual, ok := principal.GetTime("mytimestring")

	assert.True(t, ok)
	assert.Equal(t, expected, actual.Format(time.RFC3339))
}

func TestGetTimeMissingValue(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}

	actual, ok := principal.GetTime("mytime")

	assert.False(t, ok)
	assert.Equal(t, &time.Time{}, actual)
}

func TestGetTimeInvalidType(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}
	principal.Set("mywrongtype", 100)

	actual, ok := principal.GetTime("mywrongtype")

	assert.False(t, ok)
	assert.Equal(t, &time.Time{}, actual)
}

func TestGetMapOk(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}

	expected := map[string]interface{}{
		"mykey": "myvalue",
	}
	principal.Set("mymap", expected)

	actual, ok := principal.GetMap("mymap")

	assert.True(t, ok)
	assert.Equal(t, expected, actual)
}

func TestGetMapWrongType(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}

	expected := 100
	principal.Set("mywrongmap", expected)

	actual, ok := principal.GetMap("mywrongmap")

	assert.False(t, ok)
	assert.Equal(t, map[string]interface{}(nil), actual)
}

func TestGetMapExtendedTypeOk(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}

	expected := map[string]interface{}{
		"mykey": map[int]interface{}{
			0: "nested",
			1: "second",
		},
	}
	principal.Set("myextendedmap", expected)

	actual, ok := principal.GetMap("myextendedmap")

	assert.True(t, ok)
	assert.Equal(t, expected, actual)
}

func TestGetBinaryOk(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}

	expected := []byte{0, 0, 0}
	principal.Set("mybinaryarray", expected)

	actual, ok := principal.GetBinary("mybinaryarray")

	assert.True(t, ok)
	assert.Equal(t, expected, actual)
}

func TestGetBinaryMissing(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}

	actual, ok := principal.GetBinary("missingarray")

	assert.False(t, ok)
	assert.Equal(t, []byte(nil), actual)
}

func TestGetBinaryWrongType(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}
	expected := 100
	principal.Set("wrongtype", expected)

	actual, ok := principal.GetBinary("wrongtype")

	assert.False(t, ok)
	assert.Equal(t, []byte(nil), actual)
}

func TestGetURLOk(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{
		SchemaURL: url.URL{
			Scheme: "http",
			Host:   "www.example.com",
		},
	}

	actual, ok := principal.GetURL("schemaurl")

	expected, _ := url.ParseRequestURI("http://www.example.com")

	assert.True(t, ok)
	assert.Equal(t, *expected, actual)
}

func TestGetURLMissingKey(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}

	actual, ok := principal.GetURL("missing")

	assert.False(t, ok)
	assert.Equal(t, url.URL{}, actual)
}

func TestGetURLStringOk(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}
	input := "http://example.com"
	expected, _ := url.ParseRequestURI(input)

	principal.Set("stringurl", input)

	actual, ok := principal.GetURL("stringurl")

	assert.True(t, ok)
	assert.Equal(t, *expected, actual)
}

func TestGetURLStringParseErr(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}
	principal.Set("invalidurl", "")

	actual, ok := principal.GetURL("invalidurl")

	assert.False(t, ok)
	assert.Equal(t, url.URL{}, actual)
}

func TestGetURLWrongType(t *testing.T) {
	var principal cloudevents.Event = &v01.Event{}
	principal.Set("wrongtype", 100)

	actual, ok := principal.GetURL("wrongtype")

	assert.False(t, ok)
	assert.Equal(t, url.URL{}, actual)
}
func TestUnmarshalJSON(t *testing.T) {

	var actual v01.Event
	err := json.Unmarshal([]byte("{\"eventtype\":\"com.example.someevent\", \"eventtime\":\"2018-04-05T17:31:00Z\", \"myextension\":\"myValue\", \"data\": {\"topKey\" : \"topValue\", \"objectKey\": {\"embedKey\" : \"embedValue\"} }}"), &actual)
	assert.NoError(t, err)

	timestamp, _ := time.Parse(time.RFC3339, "2018-04-05T17:31:00Z")
	expected := v01.Event{
		EventType: "com.example.someevent",
		EventTime: &timestamp,
		Data: map[string]interface{}{
			"topKey": "topValue",
			"objectKey": map[string]interface{}{
				"embedKey": "embedValue",
			},
		},
	}

	expected.Set("myExtension", "myValue")
	assert.EqualValues(t, expected, actual)
}

func TestMarshallJSON(t *testing.T) {
	timestamp, _ := time.Parse(time.RFC3339, "2018-04-05T17:31:00Z")
	input := v01.Event{
		CloudEventsVersion: "0.1",
		EventID:            "1234-1234-1234",
		EventType:          "com.example.someevent",
		Source: url.URL{
			Path: "/mycontext",
		},
		EventTime: &timestamp,
		Data: map[string]interface{}{
			"topKey": "topValue",
			"objectKey": map[string]interface{}{
				"embedKey": "embedValue",
			},
		},
	}
	input.Set("myExtension", "myValue")

	actual, err := json.Marshal(input)
	expected := []byte("{\"cloudeventsversion\":\"0.1\",\"data\":{\"objectKey\":{\"embedKey\":\"embedValue\"},\"topKey\":\"topValue\"},\"eventid\":\"1234-1234-1234\",\"eventtime\":\"2018-04-05T17:31:00Z\",\"eventtype\":\"com.example.someevent\",\"myextension\":\"myValue\",\"source\":\"/mycontext\"}")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
