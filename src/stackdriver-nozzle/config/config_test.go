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

package config_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"cloud.google.com/go/compute/metadata"
	"github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/config"
)

var _ = Describe("Config", func() {

	BeforeEach(func() {
		os.Setenv("FIREHOSE_ENDPOINT", "https://api.example.com")
		os.Setenv("FIREHOSE_EVENTS_TO_STACKDRIVER_LOGGING", "LogMessage")
		os.Setenv("FIREHOSE_EVENTS_TO_STACKDRIVER_MONITORING", "")
		os.Setenv("FIREHOSE_USERNAME", "admin")
		os.Setenv("FIREHOSE_PASSWORD", "monkey123")
		os.Setenv("FIREHOSE_SKIP_SSL", "true")
		os.Setenv("FIREHOSE_SUBSCRIPTION_ID", "my-subscription-id")
		os.Setenv("FIREHOSE_NEWLINE_TOKEN", "∴")
		os.Setenv("GCP_PROJECT_ID", "test")
	})

	It("returns valid config from environment", func() {
		c, err := config.NewConfig()

		Expect(err).To(BeNil())
		Expect(c.APIEndpoint).To(Equal("https://api.example.com"))

		// Several config vals have defaults that can be overriden environment
		// that can be overriden by GCE metadata. Check those.
		funcs := []struct {
			configVal string
			localFn   func() (string, error)
			gceFn     func() (string, error)
		}{
			{c.NozzleId, func() (string, error) { return "local-nozzle", nil }, metadata.InstanceID},
			{c.NozzleName, func() (string, error) { return "local-nozzle", nil }, metadata.InstanceName},
			{c.NozzleZone, func() (string, error) { return "local-nozzle", nil }, metadata.Zone},
		}
		for _, t := range funcs {
			v, _ := t.localFn()
			if metadata.OnGCE() {
				v, _ = t.gceFn()
			}
			Expect(t.configVal).To(Equal(v))
		}

	})

	DescribeTable("required values aren't empty", func(envName string) {
		os.Setenv(envName, "")

		_, err := config.NewConfig()

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring(envName))
	},
		Entry("FIREHOSE_ENDPOINT", "FIREHOSE_ENDPOINT"),
		Entry("FIREHOSE_EVENTS_TO_STACKDRIVER_LOGGING", "FIREHOSE_EVENTS_TO_STACKDRIVER_LOGGING"),
		Entry("FIREHOSE_SUBSCRIPTION_ID", "FIREHOSE_SUBSCRIPTION_ID"),
	)
})
