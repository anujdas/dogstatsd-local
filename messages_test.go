package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDogstatsdMetricMsg(t *testing.T) {
	var tests = []struct {
		rawMsg      string
		name        string
		metricType  dogstatsdMetricType
		values      []float64
		sampleRate  float64
		tags        []string
		containerId string
	}{
		{
			"page.views:1|c",
			"page.views",
			counterMetricType,
			[]float64{1.0},
			1.0,
			[]string{},
			"",
		},
		{
			"fuel.level:0.5|g",
			"fuel.level",
			gaugeMetricType,
			[]float64{0.5},
			1.0,
			[]string{},
			"",
		},
		{
			"song.length:240|h|@0.5",
			"song.length",
			histogramMetricType,
			[]float64{240},
			0.5,
			[]string{},
			"",
		},
		{
			"users.uniques:1234|s",
			"users.uniques",
			setMetricType,
			[]float64{1234},
			1.0,
			[]string{},
			"",
		},
		{
			"users.online:1|c|@0.5|#country:china",
			"users.online",
			counterMetricType,
			[]float64{1},
			0.5,
			[]string{"country:china"},
			"",
		},
		{
			"page.views:1:2:9001|d|@0.5|#env:ci,test:1,error|c:83c0a99c0a54c0c187f461c7980e9b57f3f6a8b0c918c8d93df19a9de6f3fe1d",
			"page.views",
			distributionMetricType,
			[]float64{1, 2, 9001},
			0.5,
			[]string{"env:ci", "test:1", "error"},
			"83c0a99c0a54c0c187f461c7980e9b57f3f6a8b0c918c8d93df19a9de6f3fe1d",
		},
	}

	assert := assert.New(t)
	for _, tt := range tests {
		t.Run(tt.rawMsg, func(t *testing.T) {
			msg, _ := parseDogstatsdMsg([]byte(tt.rawMsg))
			assert.Equal(metricMsgType, msg.Type())
			assert.Equal([]byte(tt.rawMsg), msg.Data())

			metric, _ := msg.(dogstatsdMetric)
			assert.Equal(tt.name, metric.name)
			assert.Equal(tt.metricType, metric.metricType)
			assert.InDelta(time.Now().UnixMicro(), metric.ts.UnixMicro(), 100)

			floatVals := []float64{}
			for _, val := range metric.values {
				floatVals = append(floatVals, val.numeric)
			}
			assert.Equal(tt.values, floatVals)

			assert.Equal(tt.sampleRate, metric.sampleRate)
			assert.Equal(tt.tags, metric.tags)
			assert.Equal(tt.containerId, metric.containerId)

			assert.Empty(metric.extras)
		})
	}
}

func TestParseDogstatsdEventMsg(t *testing.T) {
	var tests = []struct {
		rawMsg         string
		title          string
		text           string
		ts             time.Time
		hostname       string
		aggregationKey string
		priority       dogstatsdEventPriority
		sourceType     string
		alertType      dogstatsdEventAlertType
		tags           []string
	}{
		{
			"_e{21,36}:An exception occurred|Cannot parse CSV file from 10.0.0.17|t:warning|#err_type:bad_file",
			"An exception occurred",
			"Cannot parse CSV file from 10.0.0.17",
			time.Now(),
			"",
			"",
			normalEventPriority,
			"",
			warningEventAlertType,
			[]string{"err_type:bad_file"},
		},
		{
			"_e{21,42}:An exception occurred|Cannot parse JSON request:\\\\n{\"foo: \"bar\"}|p:low|#err_type:bad_request",
			"An exception occurred",
			"Cannot parse JSON request:\\\\n{\"foo: \"bar\"}",
			time.Now(),
			"",
			"",
			lowEventPriority,
			"",
			infoEventAlertType,
			[]string{"err_type:bad_request"},
		},
		{
			"_e{5,5}:Error|Error|d:10|h:host.name|k:host.name|p:normal|s:unknown|t:error|#key:val,a:1,b",
			"Error",
			"Error",
			time.Unix(10, 0),
			"host.name",
			"host.name",
			normalEventPriority,
			"unknown",
			errorEventAlertType,
			[]string{"key:val", "a:1", "b"},
		},
	}

	assert := assert.New(t)
	for _, tt := range tests {
		t.Run(tt.rawMsg, func(t *testing.T) {
			msg, _ := parseDogstatsdMsg([]byte(tt.rawMsg))
			assert.Equal(eventMsgType, msg.Type())
			assert.Equal([]byte(tt.rawMsg), msg.Data())

			event, _ := msg.(dogstatsdEvent)
			assert.Equal(tt.title, event.title)
			assert.Equal(tt.text, event.text)
			assert.InDelta(tt.ts.UnixMicro(), event.ts.UnixMicro(), 100)
			assert.Equal(tt.hostname, event.hostname)
			assert.Equal(tt.aggregationKey, event.aggregationKey)
			assert.Equal(tt.priority, event.priority)
			assert.Equal(tt.sourceType, event.sourceType)
			assert.Equal(tt.alertType, event.alertType)
			assert.Equal(tt.tags, event.tags)
		})
	}
}

func TestParseDogstatsdServiceCheckMsg(t *testing.T) {
}
