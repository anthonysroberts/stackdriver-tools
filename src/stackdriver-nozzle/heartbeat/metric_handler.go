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

package heartbeat

import (
	"sync"
	"time"

	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/stackdriver"
	"github.com/cloudfoundry/lager"
)

type metricHandler struct {
	start      time.Time
	logger     lager.Logger
	ma         stackdriver.MetricAdapter
	nozzleId   string
	nozzleName string
	nozzleZone string

	counterMu *sync.Mutex // Guards counter
	counter   map[string]uint
}

func NewMetricHandler(ma stackdriver.MetricAdapter, logger lager.Logger, nozzleId, nozzleName, nozzleZone string) *metricHandler {
	return &metricHandler{
		logger:     logger,
		ma:         ma,
		nozzleId:   nozzleId,
		nozzleName: nozzleName,
		nozzleZone: nozzleZone,
		start:      time.Now(),
		counterMu:  &sync.Mutex{},
		counter:    map[string]uint{},
	}
}

func (h *metricHandler) Name() string {
	return "metricHandler"
}

func (h *metricHandler) Handle(event string) {
	h.counterMu.Lock()
	defer h.counterMu.Unlock()
	h.counter[event]++
	return
}

func (h *metricHandler) Flush() error {
	h.counterMu.Lock()
	defer h.counterMu.Unlock()

	metrics := []stackdriver.Metric{}
	t := time.Now()
	for k, v := range h.counter {
		metrics = append(metrics, stackdriver.Metric{
			Name:  "heartbeat." + k,
			Value: float64(v),
			Labels: map[string]string{
				"instance": h.nozzleName,
				"zone":     h.nozzleZone,
			},
			EventTime: t,
		})
	}
	h.counter = map[string]uint{}
	return h.ma.PostMetrics(metrics)
}
