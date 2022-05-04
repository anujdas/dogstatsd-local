package main

import (
	"testing"

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
			assert.Equal(msg.Type(), metricMsgType)
			assert.Equal(msg.Data(), []byte(tt.rawMsg))

			metric, _ := msg.(dogstatsdMetric)
			assert.Equal(metric.name, tt.name)
			assert.Equal(metric.metricType, tt.metricType)

			floatVals := []float64{}
			for _, val := range metric.values {
				floatVals = append(floatVals, val.numeric)
			}
			assert.Equal(floatVals, tt.values)

			assert.Equal(metric.sampleRate, tt.sampleRate)
			assert.Equal(metric.tags, tt.tags)
			assert.Equal(metric.containerId, tt.containerId)
		})
	}
}

func TestParseDogstatsdEventMsg(t *testing.T) {
}

func TestParseDogstatsdServiceCheckMsg(t *testing.T) {
}
