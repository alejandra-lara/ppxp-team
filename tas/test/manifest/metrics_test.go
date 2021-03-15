package manifest_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/planitest"
)

var _ = Describe("Metrics", func() {
	var (
		getAllInstanceGroups func(planitest.Manifest) []string
	)

	getAllInstanceGroups = func(manifest planitest.Manifest) []string {
		groups, err := manifest.Path("/instance_groups")
		Expect(err).NotTo(HaveOccurred())

		groupList, ok := groups.([]interface{})
		Expect(ok).To(BeTrue())

		names := []string{}
		for _, group := range groupList {
			groupName := group.(map[interface{}]interface{})["name"].(string)

			// ignore VMs that only contain a single placeholder job, i.e. SF-PAS only VMs that are present but non-configurable in PAS build
			jobs, err := manifest.Path(fmt.Sprintf("/instance_groups/name=%s/jobs", groupName))
			Expect(err).NotTo(HaveOccurred())
			if len(jobs.([]interface{})) > 1 {
				names = append(names, groupName)
			}
		}
		Expect(names).NotTo(BeEmpty())
		return names
	}

	Describe("prom scraper", func() {
		It("configures the prom scraper on all VMs", func() {
			manifest, err := product.RenderManifest(nil)
			Expect(err).NotTo(HaveOccurred())

			instanceGroups := getAllInstanceGroups(manifest)

			for _, ig := range instanceGroups {
				scraper, err := manifest.FindInstanceGroupJob(ig, "prom_scraper")
				Expect(err).NotTo(HaveOccurred())

				expectSecureMetrics(scraper)

				d, err := loadDomain("../../properties/metrics.yml", "prom_scraper_metrics_tls")
				Expect(err).ToNot(HaveOccurred())

				metricsProps, err := scraper.Property("metrics")
				Expect(err).ToNot(HaveOccurred())
				Expect(metricsProps).To(HaveKeyWithValue("server_name", d))
			}
		})
	})

	Describe("system metric scraper", func() {
		var instanceGroup string
		BeforeEach(func() {
			if productName == "srt" {
				instanceGroup = "control"
			} else {
				instanceGroup = "clock_global"
			}
		})

		It("configures the system-metric-scraper", func() {
			manifest, err := product.RenderManifest(nil)
			Expect(err).NotTo(HaveOccurred())

			metricScraper, err := manifest.FindInstanceGroupJob(instanceGroup, "loggr-system-metric-scraper")
			Expect(err).NotTo(HaveOccurred())

			tlsProps, err := metricScraper.Property("system_metrics/tls")
			Expect(err).ToNot(HaveOccurred())
			Expect(tlsProps).To(HaveKey("ca_cert"))
			Expect(tlsProps).To(HaveKey("cert"))
			Expect(tlsProps).To(HaveKey("key"))

			leadershipElection, err := metricScraper.Property("leadership_election")
			Expect(err).ToNot(HaveOccurred())
			Expect(leadershipElection).To(HaveKey("ca_cert"))
			Expect(leadershipElection).To(HaveKey("cert"))
			Expect(leadershipElection).To(HaveKey("key"))

			natsClient, err := metricScraper.Property("nats_client")
			Expect(err).ToNot(HaveOccurred())
			Expect(natsClient).To(HaveKey("cert"))
			Expect(natsClient).To(HaveKey("key"))

			scrapePort, err := metricScraper.Property("scrape_port")
			Expect(err).ToNot(HaveOccurred())
			Expect(scrapePort).To(Equal(53035))

			scrapePort, err = metricScraper.Property("scrape_interval")
			Expect(err).ToNot(HaveOccurred())
			Expect(scrapePort).To(Equal("1m"))

			expectSecureMetrics(metricScraper)

			d, err := loadDomain("../../properties/metrics.yml", "loggr_metric_scraper_metrics_tls")
			Expect(err).ToNot(HaveOccurred())

			metricsProps, err := metricScraper.Property("metrics")
			Expect(err).ToNot(HaveOccurred())
			Expect(metricsProps).To(HaveKeyWithValue("server_name", d))
		})

		It("has a leadership-election job collocated", func() {
			manifest, err := product.RenderManifest(nil)
			Expect(err).NotTo(HaveOccurred())

			le, err := manifest.FindInstanceGroupJob(instanceGroup, "leadership-election")
			Expect(err).NotTo(HaveOccurred())

			enabled, err := le.Property("port")
			Expect(err).ToNot(HaveOccurred())
			Expect(enabled).To(Equal(7100))
		})

		It("uses the director value of system_metrics_runtime_enabled for the enabled property", func() {
			manifest, err := product.RenderManifest(nil)
			Expect(err).NotTo(HaveOccurred())

			metricScraper, err := manifest.FindInstanceGroupJob(instanceGroup, "loggr-system-metric-scraper")
			Expect(err).NotTo(HaveOccurred())

			enabled, err := metricScraper.Property("enabled")
			Expect(err).NotTo(HaveOccurred())
			Expect(enabled).To(Equal(true))
		})
	})

	Describe("metrics discovery", func() {
		It("configures metrics-discovery on all VMs", func() {
			manifest, err := product.RenderManifest(nil)
			Expect(err).NotTo(HaveOccurred())

			instanceGroups := getAllInstanceGroups(manifest)

			for _, ig := range instanceGroups {
				registrar, err := manifest.FindInstanceGroupJob(ig, "metrics-discovery-registrar")
				Expect(err).NotTo(HaveOccurred())

				expectSecureMetrics(registrar)

				d, err := loadDomain("../../properties/metrics.yml", "metrics_discovery_metrics_tls")
				Expect(err).ToNot(HaveOccurred())

				metricsProps, err := registrar.Property("metrics")
				Expect(err).ToNot(HaveOccurred())
				Expect(metricsProps).To(HaveKeyWithValue("server_name", d))

				agent, err := manifest.FindInstanceGroupJob(ig, "metrics-agent")
				Expect(err).NotTo(HaveOccurred())

				expectSecureMetrics(agent)

				d, err = loadDomain("../../properties/metrics.yml", "metrics_agent_metrics_tls")
				Expect(err).ToNot(HaveOccurred())

				metricsProps, err = agent.Property("metrics")
				Expect(err).ToNot(HaveOccurred())
				Expect(metricsProps).To(HaveKeyWithValue("server_name", d))
			}
		})

		It("configures metrics-discovery to connect to nats locally on the instance group where nats-tls is deployed", func() {
			var natsInstanceGroupName string
			if productName == "srt" {
				natsInstanceGroupName = "database"
			} else {
				natsInstanceGroupName = "nats"
			}

			manifest, err := product.RenderManifest(nil)
			Expect(err).NotTo(HaveOccurred())

			instanceGroups := getAllInstanceGroups(manifest)
			for _, ig := range instanceGroups {
				registrar, err := manifest.FindInstanceGroupJob(ig, "metrics-discovery-registrar")
				Expect(err).NotTo(HaveOccurred())

				configuredNatsInstanceGroup, err := registrar.Property("nats_instance_group")
				Expect(err).ToNot(HaveOccurred())
				Expect(configuredNatsInstanceGroup).To(Equal(natsInstanceGroupName))
			}
		})
	})

	Describe("system metric forwarder", func() {
		var instanceGroup string
		BeforeEach(func() {
			if productName == "srt" {
				instanceGroup = "control"
			} else {
				instanceGroup = "loggregator_trafficcontroller"
			}
		})

		It("configures the legacy bosh-system-metrics-forwarder if it is enabled on the Director", func() {
			manifest, err := product.RenderManifest(nil)
			Expect(err).NotTo(HaveOccurred())

			metricsForwarder, err := manifest.FindInstanceGroupJob(instanceGroup, "bosh-system-metrics-forwarder")
			Expect(err).NotTo(HaveOccurred())

			enabled, err := metricsForwarder.Property("enabled")
			Expect(err).ToNot(HaveOccurred())
			Expect(enabled).To(Equal(true)) // This is defaulted to true in Ops Manager 2.10
		})
	})
})