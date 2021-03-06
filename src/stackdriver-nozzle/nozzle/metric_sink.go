/*
 * Copyright 2017 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nozzle

import (
	"fmt"
	"time"

	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/stackdriver"
	"github.com/cloudfoundry/sonde-go/events"
)

func NewMetricSink(labelMaker LabelMaker, metricBuffer stackdriver.MetricsBuffer, unitParser UnitParser) Sink {
	return &metricSink{
		labelMaker:   labelMaker,
		metricBuffer: metricBuffer,
		unitParser:   unitParser,
	}
}

type metricSink struct {
	labelMaker   LabelMaker
	metricBuffer stackdriver.MetricsBuffer
	unitParser   UnitParser
}

func (ms *metricSink) Receive(envelope *events.Envelope) error {
	labels := ms.labelMaker.Build(envelope)

	timestamp := time.Duration(envelope.GetTimestamp())
	eventTime := time.Unix(
		int64(timestamp/time.Second),
		int64(timestamp%time.Second),
	)

	var metrics []stackdriver.Metric
	switch envelope.GetEventType() {
	case events.Envelope_ValueMetric:
		valueMetric := envelope.GetValueMetric()
		metrics = []stackdriver.Metric{{
			Name:      valueMetric.GetName(),
			Value:     valueMetric.GetValue(),
			Labels:    labels,
			EventTime: eventTime,
			Unit:      ms.unitParser.Parse(valueMetric.GetUnit()),
		}}
	case events.Envelope_ContainerMetric:
		containerMetric := envelope.GetContainerMetric()
		metrics = []stackdriver.Metric{
			{Name: "diskBytesQuota", Value: float64(containerMetric.GetDiskBytesQuota()), EventTime: eventTime, Labels: labels},
			{Name: "instanceIndex", Value: float64(containerMetric.GetInstanceIndex()), EventTime: eventTime, Labels: labels},
			{Name: "cpuPercentage", Value: float64(containerMetric.GetCpuPercentage()), EventTime: eventTime, Labels: labels},
			{Name: "diskBytes", Value: float64(containerMetric.GetDiskBytes()), EventTime: eventTime, Labels: labels},
			{Name: "memoryBytes", Value: float64(containerMetric.GetMemoryBytes()), EventTime: eventTime, Labels: labels},
			{Name: "memoryBytesQuota", Value: float64(containerMetric.GetMemoryBytesQuota()), EventTime: eventTime, Labels: labels},
		}
	case events.Envelope_CounterEvent:
		counterEvent := envelope.GetCounterEvent()
		metrics = []stackdriver.Metric{
			{
				Name:      fmt.Sprintf("%v.delta", counterEvent.GetName()),
				Value:     float64(counterEvent.GetDelta()),
				EventTime: eventTime,
				Labels:    labels,
			},
			{
				Name:      fmt.Sprintf("%v.total", counterEvent.GetName()),
				Value:     float64(counterEvent.GetTotal()),
				EventTime: eventTime,
				Labels:    labels,
			},
		}
	default:
		return fmt.Errorf("unknown event type: %v", envelope.EventType)
	}

	for k, _ := range metrics {
		ms.metricBuffer.PostMetric(&metrics[k])
	}
	return nil
}
