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

package config

import (
	"cloud.google.com/go/compute/metadata"
	"github.com/cloudfoundry/lager"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

func NewConfig() (*Config, error) {
	var c Config
	err := envconfig.Process("", &c)
	if err != nil {
		return nil, err
	}

	err = c.validate()
	if err != nil {
		return nil, err
	}

	err = c.ensureProjectID()
	if err != nil {
		return nil, err
	}

	c.setNozzleHostInfo()

	return &c, nil
}

type Config struct {
	// Firehose config
	APIEndpoint      string `envconfig:"firehose_endpoint" required:"true"`
	LoggingEvents    string `envconfig:"firehose_events_to_stackdriver_logging" required:"true"`
	MonitoringEvents string `envconfig:"firehose_events_to_stackdriver_monitoring" required:"false"`
	Username         string `envconfig:"firehose_username" default:"admin"`
	Password         string `envconfig:"firehose_password" default:"admin"`
	SkipSSL          bool   `envconfig:"firehose_skip_ssl" default:"false"`
	SubscriptionID   string `envconfig:"firehose_subscription_id" required:"true"`
	NewlineToken     string `envconfig:"firehose_newline_token"`

	// Stackdriver config
	ProjectID             string `envconfig:"gcp_project_id"`
	MetricsBufferDuration int    `envconfig:"metrics_buffer_duration" default:"30"`
	MetricsBufferSize     int    `envconfig:"metrics_buffer_size" default:"200"`

	// Nozzle config
	HeartbeatRate      int    `envconfig:"heartbeat_rate" default:"30"`
	BatchCount         int    `envconfig:"batch_count" default:"10"`
	BatchDuration      int    `envconfig:"batch_duration" default:"1"`
	ResolveAppMetadata bool   `envconfig:"resolve_app_metadata"`
	NozzleId           string `envconfig:"nozzle_id" default:"local-nozzle"`
	NozzleName         string `envconfig:"nozzle_name" default:"local-nozzle"`
	NozzleZone         string `envconfig:"nozzle_zone" default:"local-nozzle"`
	DebugNozzle        bool   `envconfig:"debug_nozzle"`
}

func (c *Config) validate() error {
	if c.SubscriptionID == "" {
		return errors.New("FIREHOSE_SUBSCRIPTION_ID is empty")
	}

	if c.APIEndpoint == "" {
		return errors.New("FIREHOSE_ENDPOINT is empty")
	}

	if c.LoggingEvents == "" && c.MonitoringEvents == "" {
		return errors.New("FIREHOSE_EVENTS_TO_STACKDRIVER_LOGGING and FIREHOSE_EVENTS_TO_STACKDRIVER_MONITORING are empty")
	}

	return nil
}

func (c *Config) ensureProjectID() error {
	if c.ProjectID != "" {
		return nil
	}

	projectID, err := metadata.ProjectID()
	if err != nil {
		return err
	}

	c.ProjectID = projectID
	return nil
}

// If running on GCE, this will set the nozzle's ID, name, and zone to
// the GCE instance's values.
func (c *Config) setNozzleHostInfo() {
	if metadata.OnGCE() {
		if v, err := metadata.InstanceID(); err == nil {
			c.NozzleId = v
		}
		if v, err := metadata.Zone(); err == nil {
			c.NozzleZone = v
		}
		if v, err := metadata.InstanceName(); err == nil {
			c.NozzleName = v
		}
	}
}

func (c *Config) ToData() lager.Data {
	return lager.Data{
		"APIEndpoint":                   c.APIEndpoint,
		"Username":                      c.Username,
		"Password":                      "<redacted>",
		"EventsToStackdriverMonitoring": c.MonitoringEvents,
		"EventsToStackdriverLogging":    c.LoggingEvents,
		"SkipSSL":                       c.SkipSSL,
		"ProjectID":                     c.ProjectID,
		"BatchCount":                    c.BatchCount,
		"BatchDuration":                 c.BatchDuration,
		"HeartbeatRate":                 c.HeartbeatRate,
		"ResolveAppMetadata":            c.ResolveAppMetadata,
		"SubscriptionID":                c.SubscriptionID,
		"DebugNozzle":                   c.DebugNozzle,
		"NewlineToken":                  c.NewlineToken,
	}
}
