/*
Copyright (c) 2016-2017 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package utils

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/Shopify/sarama"
)

func GenConsumerMessageWithBody(body string) *sarama.ConsumerMessage {
	timestamp, err := time.Parse(time.RFC3339, "2020-11-27T14:41:29+02:00")
	if err != nil {
		fmt.Println("Unable to parse time")
	}
	return &sarama.ConsumerMessage{
		Offset:    1023435314301,
		Partition: 2,
		Topic:     "mytopic",
		Value:     []byte(body),
		Key:       []byte("1234"),
		Timestamp: timestamp,
	}
}
func TestGetHTTPRequest(t *testing.T) {
	value := "my msg"
	msg := GenConsumerMessageWithBody(value)
	req, err := GetHTTPReq("foo", 1234, msg, "myns", "kafkatriggers.kubeless.io")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if req.Method != "POST" {
		t.Errorf("Unexpected method %s", req.Method)
	}
	if req.URL.String() != "http://foo.myns.svc.cluster.local:1234" {
		t.Errorf("Unexpected URL %s", req.URL.String())
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if string(body) != "my msg" {
		t.Errorf("Unexpected method %s", string(body))
	}
	if req.Header.Get("event-id") == "" {
		t.Error("Header event-id should be set")
	}
	if req.Header.Get("event-time") == "" {
		t.Error("Header event-time should be set")
	}
	if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		t.Errorf("Unexpected Content-Type %s", req.Header.Get("Content-Type"))
	}
	if req.Header.Get("event-type") != "application/x-www-form-urlencoded" {
		t.Errorf("Unexpected event-type %s", req.Header.Get("event-type"))
	}
	if req.Header.Get("event-namespace") != "kafkatriggers.kubeless.io" {
		t.Errorf("Unexpected event-type %s", req.Header.Get("event-type"))
	}
	if req.Header.Get("event-topic") != "mytopic" {
		t.Errorf("Unexpected event-topic %s", req.Header.Get("event-topic"))
	}
	if req.Header.Get("event-offset") != "1023435314301" {
		t.Errorf("Unexpected event-offset %s", req.Header.Get("event-offset"))
	}
	if req.Header.Get("event-message-timestamp") != "2020-11-27T12:41:29Z" {
		t.Errorf("Unexpected event-message-timestamp %s", req.Header.Get("event-message-timestamp"))
	}
	if req.Header.Get("event-key") != "1234" {
		t.Errorf("Unexpected event-key %s", req.Header.Get("event-key"))
	}
}

func TestGetJSONHTTPRequest(t *testing.T) {
	value := `{"hello": "world"}`
	msg := GenConsumerMessageWithBody(value)
	req, err := GetHTTPReq("foo", int(1234), msg, "myns", "kafkatriggers.kubeless.io")
	// req, err := GetHTTPReq("foo", 1234, msg, "myns", "kafkatriggers.kubeless.io", "POST", `{"hello": "world"}`)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Unexpected Content-Type %s", req.Header.Get("Content-Type"))
	}
	if req.Header.Get("event-type") != "application/json" {
		t.Errorf("Unexpected event-type %s", req.Header.Get("event-type"))
	}
	if req.Header.Get("event-topic") != "mytopic" {
		t.Errorf("Unexpected event-topic %s", req.Header.Get("event-topic"))
	}
}

func TestIsJSON(t *testing.T) {
	type testObj struct {
		input        string
		expectedJSON bool
	}
	testData := []testObj{
		{
			input:        `{"A": "B"}`,
			expectedJSON: true,
		},
		{
			input:        `{"hello": "World"}`,
			expectedJSON: true,
		},
		{
			input: `{
				"hello": "World"
			}`,
			expectedJSON: true,
		},
		{
			input:        "{\n\"hello\": \"World\"\n}",
			expectedJSON: true,
		},
		{
			input:        `{"A": "B"`,
			expectedJSON: false,
		},
		{
			input:        `hello world`,
			expectedJSON: false,
		},
	}
	for _, d := range testData {
		if IsJSON(d.input) != d.expectedJSON {
			t.Errorf("isJSON(%s) should be %t", d.input, d.expectedJSON)
		}
	}
}
