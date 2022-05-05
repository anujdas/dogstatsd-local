package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type dogstatsdMsgType int

func (d dogstatsdMsgType) String() string {
	switch d {
	case metricMsgType:
		return "metric"
	case eventMsgType:
		return "event"
	case serviceCheckMsgType:
		return "service_check"
	}

	return "unknown"
}

const (
	metricMsgType dogstatsdMsgType = iota
	serviceCheckMsgType
	eventMsgType
)

type dogstatsdMsg interface {
	Type() dogstatsdMsgType
	Data() []byte
}

func parseDogstatsdMetricMsg(buf []byte) (dogstatsdMsg, error) {
	metric := dogstatsdMetric{
		data:       buf,
		ts:         time.Now(),
		values:     []dogstatsdMetricValue{},
		sampleRate: 1.0,
		tags:       []string{},
	}

	// sample message: metric.name:value1:value2|type|@sample_rate|#tag1:value,tag2|c:container_id
	pieces := strings.Split(string(buf), "|")
	if len(pieces) < 2 {
		return nil, errors.New("INVALID_MSG_MISSING_NAME_VALUE_OR_TYPE")
	}

	addrAndValues := strings.Split(pieces[0], ":")
	if len(addrAndValues) < 2 {
		return nil, fmt.Errorf("INVALID_MSG_MISSING_NAME_AND_VALUE (%s)", pieces[0])
	}

	metric.name = addrAndValues[0]
	rawValues := addrAndValues[1:]

	switch pieces[1] {
	case "c":
		metric.metricType = counterMetricType
	case "g":
		metric.metricType = gaugeMetricType
	case "s":
		metric.metricType = setMetricType
	case "ms":
		metric.metricType = timerMetricType
	case "h":
		metric.metricType = histogramMetricType
	case "d":
		metric.metricType = distributionMetricType
	default:
		return nil, fmt.Errorf("INVALID_MSG_INVALID_TYPE (%s)", pieces[1])
	}

	// all numeric values are ints or floats, stored as floats
	for _, rawValue := range rawValues {
		value := dogstatsdMetricValue{
			raw: rawValue,
		}

		floatValue, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return nil, fmt.Errorf("INVALID_MSG_INVALID_VALUE (%s)", rawValue)
		}
		value.numeric = floatValue

		if metric.metricType == timerMetricType {
			value.duration = time.Duration(value.numeric) / time.Millisecond
		}

		metric.values = append(metric.values, value)
	}

	// parse out sample rate, tags, container id, and any extras
	for _, piece := range pieces[2:] {
		if strings.HasPrefix(piece, "@") {
			sampleRate, err := strconv.ParseFloat(piece[1:], 64)
			if err != nil {
				return nil, fmt.Errorf("INVALID_SAMPLE_RATE (%s)", piece[:1])
			}
			metric.sampleRate = sampleRate
			continue
		}

		if strings.HasPrefix(piece, "#") {
			tags := strings.Split(piece[1:], ",")
			metric.tags = append(metric.tags, tags...)
			continue
		}

		if strings.HasPrefix(piece, "c:") {
			metric.containerId = piece[2:]
			continue
		}

		metric.extras = append(metric.extras, piece)
	}

	return metric, nil
}

type dogstatsdMetricType int

func (d dogstatsdMetricType) String() string {
	switch d {
	case gaugeMetricType:
		return "gauge"
	case counterMetricType:
		return "counter"
	case setMetricType:
		return "set"
	case timerMetricType:
		return "timer"
	case histogramMetricType:
		return "histogram"
	case distributionMetricType:
		return "distribution"
	}

	return "unknown"
}

const (
	gaugeMetricType dogstatsdMetricType = iota
	counterMetricType
	setMetricType
	timerMetricType
	histogramMetricType
	distributionMetricType
)

type dogstatsdMetricValue struct {
	raw      string
	numeric  float64
	duration time.Duration
}

type dogstatsdMetric struct {
	data []byte
	ts   time.Time

	name string

	metricType dogstatsdMetricType
	values     []dogstatsdMetricValue

	sampleRate  float64
	tags        []string
	containerId string
	extras      []string
}

func (d dogstatsdMetric) Data() []byte {
	return d.data
}

func (d dogstatsdMetric) Type() dogstatsdMsgType {
	return metricMsgType
}

// _sc|<NAME>|<STATUS>|d:<TIMESTAMP>|h:<HOSTNAME>|#<TAG_KEY_1>:<TAG_VALUE_1>,<TAG_2>|m:<SERVICE_CHECK_MESSAGE>
func parseDogstatsdServiceCheckMsg(buf []byte) (dogstatsdMsg, error) {
	serviceCheck := dogstatsdServiceCheck{
		data:   buf,
		ts:     time.Now(),
		tags:   []string{},
		extras: []string{},
	}

	pieces := strings.Split(string(buf), "|")
	if len(pieces) < 3 {
		return nil, errors.New("INVALID_MSG_MISSING_NAME_OR_STATUS")
	}

	serviceCheck.name = pieces[1]

	switch pieces[2] {
	case "0":
		serviceCheck.status = okServiceCheckStatusType
	case "1":
		serviceCheck.status = warningServiceCheckStatusType
	case "2":
		serviceCheck.status = criticalServiceCheckStatusType
	case "3":
		serviceCheck.status = unknownServiceCheckStatusType
	default:
		return nil, fmt.Errorf("INVALID_MSG_INVALID_STATUS (%s)", pieces[2])
	}

	for _, piece := range pieces[3:] {
		if strings.HasPrefix(piece, "d:") {
			unixTime, err := strconv.ParseInt(piece[2:], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("INVALID_TIMESTAMP (%s)", piece[2:])
			}

			serviceCheck.ts = time.Unix(unixTime, 0)
			continue
		}

		if strings.HasPrefix(piece, "h:") {
			serviceCheck.hostname = piece[2:]
			continue
		}

		if strings.HasPrefix(piece, "#") {
			tags := strings.Split(piece[1:], ",")
			serviceCheck.tags = append(serviceCheck.tags, tags...)
			continue
		}

		if strings.HasPrefix(piece, "m:") {
			serviceCheck.message = piece[2:]
			continue
		}

		serviceCheck.extras = append(serviceCheck.extras, piece)
	}

	return serviceCheck, nil
}

type dogstatsdServiceCheckStatus int

const (
	okServiceCheckStatusType dogstatsdServiceCheckStatus = iota
	warningServiceCheckStatusType
	criticalServiceCheckStatusType
	unknownServiceCheckStatusType
)

func (s dogstatsdServiceCheckStatus) String() string {
	switch s {
	case okServiceCheckStatusType:
		return "OK"
	case warningServiceCheckStatusType:
		return "WARNING"
	case criticalServiceCheckStatusType:
		return "CRITICAL"
	}
	return "UNKNOWN"
}

type dogstatsdServiceCheck struct {
	data     []byte
	name     string
	status   dogstatsdServiceCheckStatus
	ts       time.Time
	hostname string
	tags     []string
	message  string
	extras   []string
}

func (dogstatsdServiceCheck) Type() dogstatsdMsgType {
	return serviceCheckMsgType
}

func (d dogstatsdServiceCheck) Data() []byte {
	return d.data
}

// docs: https://docs.datadoghq.com/developers/dogstatsd/datagram_shell/?tab=events
// _e{<TITLE_UTF8_LENGTH>,<TEXT_UTF8_LENGTH>}:<TITLE>|<TEXT>|d:<TIMESTAMP>|h:<HOSTNAME>|p:<PRIORITY>|t:<ALERT_TYPE>|#<TAG_KEY_1>:<TAG_VALUE_1>,<TAG_2>
func parseDogstatsdEventMsg(buf []byte) (dogstatsdMsg, error) {
	event := dogstatsdEvent{
		data: buf,
		ts:   time.Now(),
		tags: []string{},
	}

	pieces := strings.Split(string(buf), "|")
	if len(pieces) < 2 {
		return nil, errors.New("INVALID_MSG_MISSING_TITLE_OR_TEXT")
	}

	lengthsAndTitle := strings.Split(pieces[0], ":")
	if len(lengthsAndTitle) != 2 {
		return nil, fmt.Errorf("INVALID_MSG_MISSING_TITLE (%s)", pieces[0])
	}
	event.title = lengthsAndTitle[1]
	event.text = pieces[1]

	for _, piece := range pieces[2:] {
		if strings.HasPrefix(piece, "d:") {
			unixTime, err := strconv.ParseInt(piece[2:], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("INVALID_TIMESTAMP (%s)", piece[2:])
			}

			event.ts = time.Unix(unixTime, 0)
			continue
		}

		if strings.HasPrefix(piece, "h:") {
			event.hostname = piece[2:]
			continue
		}

		if strings.HasPrefix(piece, "k:") {
			event.aggregationKey = piece[2:]
			continue
		}

		if strings.HasPrefix(piece, "p:") {
			switch piece[2:] {
			case "low":
				event.priority = lowEventPriority
			case "normal":
				event.priority = normalEventPriority
			default:
				return nil, fmt.Errorf("INVALID_MSG_INVALID_PRIORITY (%s)", piece[2:])
			}

			continue
		}

		if strings.HasPrefix(piece, "s:") {
			event.sourceType = piece[2:]
			continue
		}

		if strings.HasPrefix(piece, "t:") {
			switch piece[2:] {
			case "info":
				event.alertType = infoEventAlertType
			case "success":
				event.alertType = successEventAlertType
			case "warning":
				event.alertType = warningEventAlertType
			case "error":
				event.alertType = errorEventAlertType
			default:
				return nil, fmt.Errorf("INVALID_MSG_INVALID_ALERT_TYPE (%s)", piece[2:])
			}

			continue
		}

		if strings.HasPrefix(piece, "#") {
			tags := strings.Split(piece[1:], ",")
			event.tags = append(event.tags, tags...)
			continue
		}

		event.extras = append(event.extras, piece)
	}

	return event, nil
}

type dogstatsdEventPriority int

const (
	normalEventPriority dogstatsdEventPriority = iota
	lowEventPriority
)

func (p dogstatsdEventPriority) String() string {
	switch p {
	case normalEventPriority:
		return "normal"
	case lowEventPriority:
		return "low"
	}
	return "unknown"
}

type dogstatsdEventAlertType int

const (
	infoEventAlertType dogstatsdEventAlertType = iota
	successEventAlertType
	warningEventAlertType
	errorEventAlertType
)

func (a dogstatsdEventAlertType) String() string {
	switch a {
	case infoEventAlertType:
		return "info"
	case successEventAlertType:
		return "success"
	case warningEventAlertType:
		return "warning"
	case errorEventAlertType:
		return "error"
	}
	return "unknown"
}

type dogstatsdEvent struct {
	data  []byte
	title string
	text  string

	ts             time.Time
	hostname       string
	aggregationKey string
	priority       dogstatsdEventPriority
	sourceType     string
	alertType      dogstatsdEventAlertType
	tags           []string
	extras         []string
}

func (e dogstatsdEvent) Data() []byte {
	return e.data
}

func (e dogstatsdEvent) Type() dogstatsdMsgType {
	return eventMsgType
}

// parse a dogstatsdMsg, returning the correct message back
func parseDogstatsdMsg(buf []byte) (dogstatsdMsg, error) {
	if bytes.HasPrefix(buf, []byte("_e{")) {
		return parseDogstatsdEventMsg(buf)
	}

	if bytes.HasPrefix(buf, []byte("_sc")) {
		return parseDogstatsdServiceCheckMsg(buf)
	}

	return parseDogstatsdMetricMsg(buf)
}
