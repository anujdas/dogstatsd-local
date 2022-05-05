package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type dogstatsdJsonMetric struct {
	Name string `json:"name"`
	Type string `json:"type"`

	Values      []float64 `json:"values"`
	SampleRate  float64   `json:"sample_rate"`
	Tags        []string  `json:"tags"`
	ContainerId string    `json:"container_id"`
}

func newJsonDogstatsdMsgHandler() msgHandler {
	return func(msg []byte) error {
		dMsg, err := parseDogstatsdMsg(msg)
		if err != nil {
			log.Println(err.Error())
		}

		if dMsg.Type() != metricMsgType {
			log.Println("Unable to serialize non metric messages to JSON yet")
			return nil
		}

		metric, ok := dMsg.(dogstatsdMetric)
		if !ok {
			log.Fatalf("Programming error: invalid Type() = type matching")
		}

		floatValues := make([]float64, 0)
		for _, value := range metric.values {
			floatValues = append(floatValues, value.numeric)
		}

		jsonMsg := dogstatsdJsonMetric{
			Name:        metric.name,
			Type:        metric.metricType.String(),
			Values:      floatValues,
			SampleRate:  metric.sampleRate,
			Tags:        metric.tags,
			ContainerId: metric.containerId,
		}

		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(&jsonMsg); err != nil {
			log.Println("JSON serialize error:", err.Error())
			return nil
		}

		return nil
	}
}

func newHumanDogstatsdMsgHandler() msgHandler {
	return func(msg []byte) error {
		dMsg, err := parseDogstatsdMsg(msg)
		if err != nil {
			log.Println(err.Error())
			return nil
		}

		metric, _ := dMsg.(dogstatsdMetric)
		if dMsg.Type() != metricMsgType {
			fmt.Println(string(dMsg.Data()))
			return nil
		}

		values := make([]string, 0)
		for _, value := range metric.values {
			strValue := fmt.Sprintf("%.2f", value.numeric)
			if metric.metricType == timerMetricType {
				strValue += "ms"
			}

			values = append(values, strValue)
		}

		str := fmt.Sprintf(
			"metric:%s|%s|%s %s",
			metric.metricType.String(),
			metric.name,
			strings.Join(values, ","),
			strings.Join(metric.tags, " "),
		)

		fmt.Println(str)

		return nil
	}
}

func newRawDogstatsdMsgHandler() msgHandler {
	return func(msg []byte) error {
		fmt.Fprintf(os.Stdout, string(msg))
		fmt.Fprintf(os.Stdout, "\n")
		return nil
	}
}
