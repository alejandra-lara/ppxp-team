package manifest_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Networking", func() {
	Describe("Container networking", func() {
		var (
			inputProperties         map[string]interface{}
			cellInstanceGroup       string
			controllerInstanceGroup string
		)

		BeforeEach(func() {
			if productName == "ert" {
				cellInstanceGroup = "diego_cell"
				controllerInstanceGroup = "diego_database"
			} else {
				cellInstanceGroup = "compute"
				controllerInstanceGroup = "control"
			}
			inputProperties = map[string]interface{}{}
		})

		Describe("policy server", func() {
			It("uses the correct database host", func() {
				manifest, err := product.RenderManifest(inputProperties)
				Expect(err).NotTo(HaveOccurred())

				job, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "policy-server")
				Expect(err).NotTo(HaveOccurred())

				host, err := job.Property("database/host")
				Expect(err).NotTo(HaveOccurred())
				Expect(host).To(Equal("mysql.service.cf.internal"))

				databaseLink, err := job.Path("/consumes/database")
				Expect(err).NotTo(HaveOccurred())
				Expect(databaseLink).To(Equal("nil"))
			})

			Context("when the operator does not set a limit for policy server open database connections", func() {
				It("configures jobs with default values", func() {
					manifest, err := product.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "policy-server")
					Expect(err).NotTo(HaveOccurred())

					maxOpenConnections, err := job.Property("max_open_connections")
					Expect(err).NotTo(HaveOccurred())
					Expect(maxOpenConnections).To(Equal(200))
				})
			})

			Context("when the user specifies custom values for policy server max open database connections", func() {
				BeforeEach(func() {
					inputProperties = map[string]interface{}{
						".properties.networkpolicyserver_database_max_open_connections": 300,
					}
				})

				It("configures jobs with user provided values", func() {
					manifest, err := product.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "policy-server")
					Expect(err).NotTo(HaveOccurred())

					maxOpenConnections, err := job.Property("max_open_connections")
					Expect(err).NotTo(HaveOccurred())
					Expect(maxOpenConnections).To(Equal(300))
				})

				Context("when the policy server max open DB connections is out of range", func() {
					BeforeEach(func() {
						inputProperties = map[string]interface{}{
							".properties.networkpolicyserver_database_max_open_connections": 0,
						}
					})

					It("returns an error", func() {
						_, err := product.RenderManifest(inputProperties)
						Expect(err.Error()).To(ContainSubstring("Value must be greater than or equal to 1"))
					})
				})
			})

			It("disables TLS to the internal database by default", func() {
				manifest, err := product.RenderManifest(inputProperties)
				Expect(err).NotTo(HaveOccurred())

				job, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "policy-server")
				Expect(err).NotTo(HaveOccurred())

				tlsEnabled, err := job.Property("database/require_ssl")
				Expect(err).NotTo(HaveOccurred())
				Expect(tlsEnabled).To(BeFalse())

				caCert, err := job.Property("database/ca_cert")
				Expect(err).NotTo(HaveOccurred())
				Expect(caCert).To(BeNil())
			})

			Context("when the TLS checkbox is checked", func() {
				BeforeEach(func() {
					inputProperties = map[string]interface{}{".properties.enable_tls_to_internal_pxc": true}
				})

				It("configures TLS to the internal database", func() {
					manifest, err := product.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "policy-server")
					Expect(err).NotTo(HaveOccurred())

					tlsEnabled, err := job.Property("database/require_ssl")
					Expect(err).NotTo(HaveOccurred())
					Expect(tlsEnabled).To(BeTrue())

					caCert, err := job.Property("database/ca_cert")
					Expect(err).NotTo(HaveOccurred())
					Expect(caCert).NotTo(BeEmpty())
				})
			})
			When("the system database is set to internal", func() {
				It("does not skip hostname validation", func() {
					manifest, renderErr := product.RenderManifest(nil)
					Expect(renderErr).NotTo(HaveOccurred())

					routing_api, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "policy-server")
					Expect(err).ToNot(HaveOccurred())
					skip_hostname_validation, err := routing_api.Property("database/skip_hostname_validation")
					Expect(err).ToNot(HaveOccurred())
					Expect(skip_hostname_validation).To(BeFalse())
				})
			})

			When("the system database is set to external", func() {
				var inputProperties map[string]interface{}
				BeforeEach(func() {
					inputProperties = map[string]interface{}{
						".properties.system_database":                                       "external",
						".properties.system_database.external.host":                         "foo.bar",
						".properties.system_database.external.validate_hostname":            false,
						".properties.system_database.external.port":                         5432,
						".properties.system_database.external.credhub_username":             "some-user",
						".properties.system_database.external.credhub_password":             map[string]interface{}{"secret": "some-password"},
						".properties.system_database.external.app_usage_service_username":   "app_usage_service_username",
						".properties.system_database.external.app_usage_service_password":   map[string]interface{}{"secret": "app_usage_service_password"},
						".properties.system_database.external.autoscale_username":           "autoscale_username",
						".properties.system_database.external.autoscale_password":           map[string]interface{}{"secret": "autoscale_password"},
						".properties.system_database.external.ccdb_username":                "ccdb_username",
						".properties.system_database.external.ccdb_password":                map[string]interface{}{"secret": "ccdb_password"},
						".properties.system_database.external.diego_username":               "diego_username",
						".properties.system_database.external.diego_password":               map[string]interface{}{"secret": "diego_password"},
						".properties.system_database.external.locket_username":              "locket_username",
						".properties.system_database.external.locket_password":              map[string]interface{}{"secret": "locket_password"},
						".properties.system_database.external.networkpolicyserver_username": "networkpolicyserver_username",
						".properties.system_database.external.networkpolicyserver_password": map[string]interface{}{"secret": "networkpolicyserver_password"},
						".properties.system_database.external.notifications_username":       "notifications_username",
						".properties.system_database.external.notifications_password":       map[string]interface{}{"secret": "notifications_password"},
						".properties.system_database.external.account_username":             "account_username",
						".properties.system_database.external.account_password":             map[string]interface{}{"secret": "account_password"},
						".properties.system_database.external.routing_username":             "routing_username",
						".properties.system_database.external.routing_password":             map[string]interface{}{"secret": "routing_password"},
						".properties.system_database.external.silk_username":                "silk_username",
						".properties.system_database.external.silk_password":                map[string]interface{}{"secret": "silk_password"},
					}
				})

				It("skips hostname validation", func() {
					manifest, renderErr := product.RenderManifest(inputProperties)
					Expect(renderErr).NotTo(HaveOccurred())

					routing_api, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "policy-server")
					Expect(err).ToNot(HaveOccurred())
					skip_hostname_validation, err := routing_api.Property("database/skip_hostname_validation")
					Expect(err).ToNot(HaveOccurred())
					Expect(skip_hostname_validation).To(BeTrue())
				})
			})
		})

		Describe("policy server internal", func() {
			Context("when the operator does not set a limit for policy server internal open database connections", func() {
				It("configures jobs with default values", func() {
					manifest, err := product.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "policy-server-internal")
					Expect(err).NotTo(HaveOccurred())

					maxOpenConnections, err := job.Property("max_open_connections")
					Expect(err).NotTo(HaveOccurred())
					Expect(maxOpenConnections).To(Equal(200))
				})
			})

			Context("when the user specifies custom values for policy server internal max open database connections", func() {
				BeforeEach(func() {
					inputProperties = map[string]interface{}{
						".properties.networkpolicyserverinternal_database_max_open_connections": 300,
					}
				})

				It("configures jobs with user provided values", func() {
					manifest, err := product.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "policy-server-internal")
					Expect(err).NotTo(HaveOccurred())

					maxOpenConnections, err := job.Property("max_open_connections")
					Expect(err).NotTo(HaveOccurred())
					Expect(maxOpenConnections).To(Equal(300))
				})

				Context("when the policy server internal max open DB connections is out of range", func() {
					BeforeEach(func() {
						inputProperties = map[string]interface{}{
							".properties.networkpolicyserverinternal_database_max_open_connections": 0,
						}
					})

					It("returns an error", func() {
						_, err := product.RenderManifest(inputProperties)
						Expect(err.Error()).To(ContainSubstring("Value must be greater than or equal to 1"))
					})
				})
			})
		})

		Context("when the operator configures database connection timeout for CNI plugin", func() {
			BeforeEach(func() {
				inputProperties = map[string]interface{}{
					".properties.cf_networking_database_connection_timeout": 250,
				}
			})

			It("sets the manifest database connection timeout properties for the cf networking jobs to be 250", func() {
				manifest, err := product.RenderManifest(inputProperties)
				Expect(err).NotTo(HaveOccurred())

				policyServerJob, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "policy-server")
				Expect(err).NotTo(HaveOccurred())

				policyServerConnectTimeoutSeconds, err := policyServerJob.Property("database/connect_timeout_seconds")
				Expect(err).NotTo(HaveOccurred())

				Expect(policyServerConnectTimeoutSeconds).To(Equal(250))

				policyServerInternalJob, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "policy-server-internal")
				Expect(err).NotTo(HaveOccurred())

				policyServerInternalConnectTimeoutSeconds, err := policyServerInternalJob.Property("database/connect_timeout_seconds")
				Expect(err).NotTo(HaveOccurred())

				Expect(policyServerInternalConnectTimeoutSeconds).To(Equal(250))

				silkControllerJob, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "silk-controller")
				Expect(err).NotTo(HaveOccurred())

				silkControllerConnectTimeoutSeconds, err := silkControllerJob.Property("database/connect_timeout_seconds")
				Expect(err).NotTo(HaveOccurred())

				Expect(silkControllerConnectTimeoutSeconds).To(Equal(250))
			})
		})

		Context("when Silk is enabled", func() {
			BeforeEach(func() {
				inputProperties = map[string]interface{}{}
			})

			It("configures the cni_config_dir and cni_plugin_dir", func() {
				manifest, err := product.RenderManifest(inputProperties)
				Expect(err).NotTo(HaveOccurred())

				job, err := manifest.FindInstanceGroupJob(cellInstanceGroup, "garden-cni")
				Expect(err).NotTo(HaveOccurred())

				cniConfigDir, err := job.Property("cni_config_dir")
				Expect(err).NotTo(HaveOccurred())
				Expect(cniConfigDir).To(Equal("/var/vcap/jobs/silk-cni/config/cni"))

				cniPluginDir, err := job.Property("cni_plugin_dir")
				Expect(err).NotTo(HaveOccurred())
				Expect(cniPluginDir).To(Equal("/var/vcap/packages/silk-cni/bin"))
			})

			It("disables TLS to the internal database by default", func() {
				manifest, err := product.RenderManifest(inputProperties)
				Expect(err).NotTo(HaveOccurred())

				job, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "silk-controller")
				Expect(err).NotTo(HaveOccurred())

				tlsEnabled, err := job.Property("database/require_ssl")
				Expect(err).NotTo(HaveOccurred())
				Expect(tlsEnabled).To(BeFalse())

				caCert, err := job.Property("database/ca_cert")
				Expect(err).NotTo(HaveOccurred())
				Expect(caCert).To(BeNil())
			})

			Context("when the TLS checkbox is checked", func() {
				BeforeEach(func() {
					inputProperties = map[string]interface{}{".properties.enable_tls_to_internal_pxc": true}
				})

				It("configures TLS to the internal database", func() {
					manifest, err := product.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "silk-controller")
					Expect(err).NotTo(HaveOccurred())

					tlsEnabled, err := job.Property("database/require_ssl")
					Expect(err).NotTo(HaveOccurred())
					Expect(tlsEnabled).To(BeTrue())

					caCert, err := job.Property("database/ca_cert")
					Expect(err).NotTo(HaveOccurred())
					Expect(caCert).NotTo(BeEmpty())
				})
			})

			It("uses the correct database host", func() {
				manifest, err := product.RenderManifest(inputProperties)
				Expect(err).NotTo(HaveOccurred())

				job, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "silk-controller")
				Expect(err).NotTo(HaveOccurred())

				host, err := job.Property("database/host")
				Expect(err).NotTo(HaveOccurred())
				Expect(host).To(Equal("mysql.service.cf.internal"))

				databaseLink, err := job.Path("/consumes/database")
				Expect(err).NotTo(HaveOccurred())
				Expect(databaseLink).To(Equal("nil"))
			})

			When("the database is set to internal", func() {
				It("does not skip hostname validation", func() {
					manifest, renderErr := product.RenderManifest(nil)
					Expect(renderErr).NotTo(HaveOccurred())

					routing_api, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "silk-controller")
					Expect(err).ToNot(HaveOccurred())
					skip_hostname_validation, err := routing_api.Property("database/skip_hostname_validation")
					Expect(err).ToNot(HaveOccurred())
					Expect(skip_hostname_validation).To(BeFalse())
				})
			})

			When("the database is set to external", func() {
				var inputProperties map[string]interface{}
				BeforeEach(func() {
					inputProperties = map[string]interface{}{
						".properties.system_database":                                       "external",
						".properties.system_database.external.host":                         "foo.bar",
						".properties.system_database.external.validate_hostname":            false,
						".properties.system_database.external.port":                         5432,
						".properties.system_database.external.credhub_username":             "some-user",
						".properties.system_database.external.credhub_password":             map[string]interface{}{"secret": "some-password"},
						".properties.system_database.external.app_usage_service_username":   "app_usage_service_username",
						".properties.system_database.external.app_usage_service_password":   map[string]interface{}{"secret": "app_usage_service_password"},
						".properties.system_database.external.autoscale_username":           "autoscale_username",
						".properties.system_database.external.autoscale_password":           map[string]interface{}{"secret": "autoscale_password"},
						".properties.system_database.external.ccdb_username":                "ccdb_username",
						".properties.system_database.external.ccdb_password":                map[string]interface{}{"secret": "ccdb_password"},
						".properties.system_database.external.diego_username":               "diego_username",
						".properties.system_database.external.diego_password":               map[string]interface{}{"secret": "diego_password"},
						".properties.system_database.external.locket_username":              "locket_username",
						".properties.system_database.external.locket_password":              map[string]interface{}{"secret": "locket_password"},
						".properties.system_database.external.networkpolicyserver_username": "networkpolicyserver_username",
						".properties.system_database.external.networkpolicyserver_password": map[string]interface{}{"secret": "networkpolicyserver_password"},
						".properties.system_database.external.notifications_username":       "notifications_username",
						".properties.system_database.external.notifications_password":       map[string]interface{}{"secret": "notifications_password"},
						".properties.system_database.external.account_username":             "account_username",
						".properties.system_database.external.account_password":             map[string]interface{}{"secret": "account_password"},
						".properties.system_database.external.routing_username":             "routing_username",
						".properties.system_database.external.routing_password":             map[string]interface{}{"secret": "routing_password"},
						".properties.system_database.external.silk_username":                "silk_username",
						".properties.system_database.external.silk_password":                map[string]interface{}{"secret": "silk_password"},
					}
				})

				It("skips hostname validation", func() {
					manifest, renderErr := product.RenderManifest(inputProperties)
					Expect(renderErr).NotTo(HaveOccurred())

					routing_api, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "silk-controller")
					Expect(err).ToNot(HaveOccurred())
					skip_hostname_validation, err := routing_api.Property("database/skip_hostname_validation")
					Expect(err).ToNot(HaveOccurred())
					Expect(skip_hostname_validation).To(BeTrue())
				})
			})

			Context("when the operator does not set a limit for silk-controller open database connections", func() {
				It("configures jobs with default values", func() {
					manifest, err := product.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "silk-controller")
					Expect(err).NotTo(HaveOccurred())

					maxOpenConnections, err := job.Property("max_open_connections")
					Expect(err).NotTo(HaveOccurred())
					Expect(maxOpenConnections).To(Equal(200))
				})
			})

			Context("when the user specifies custom values for silk-controller max open database connections", func() {
				BeforeEach(func() {
					inputProperties = map[string]interface{}{
						".properties.silk_database_max_open_connections": 300,
					}
				})

				It("configures jobs with user provided values", func() {
					manifest, err := product.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(controllerInstanceGroup, "silk-controller")
					Expect(err).NotTo(HaveOccurred())

					maxOpenConnections, err := job.Property("max_open_connections")
					Expect(err).NotTo(HaveOccurred())
					Expect(maxOpenConnections).To(Equal(300))
				})

				Context("when the silk-controller max open DB connections is out of range", func() {
					BeforeEach(func() {
						inputProperties = map[string]interface{}{
							".properties.silk_database_max_open_connections": 0,
						}
					})

					It("returns an error", func() {
						_, err := product.RenderManifest(inputProperties)
						Expect(err.Error()).To(ContainSubstring("Value must be greater than or equal to 1"))
					})
				})
			})

			Context("silk network policy", func() {
				It("continues to be enforced", func() {
					manifest, err := product.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(cellInstanceGroup, "vxlan-policy-agent")
					Expect(err).NotTo(HaveOccurred())

					disabled, err := job.Property("disable_container_network_policy")
					Expect(err).NotTo(HaveOccurred())
					Expect(disabled).To(BeFalse())
				})

				Context("setting is disabled", func() {
					BeforeEach(func() {
						inputProperties = map[string]interface{}{
							".properties.container_networking_interface_plugin.silk.enable_policy_enforcement": false,
						}
					})

					It("disables silk network policy enforcement", func() {
						manifest, err := product.RenderManifest(inputProperties)
						Expect(err).NotTo(HaveOccurred())

						job, err := manifest.FindInstanceGroupJob(cellInstanceGroup, "vxlan-policy-agent")
						Expect(err).NotTo(HaveOccurred())

						disabled, err := job.Property("disable_container_network_policy")
						Expect(err).NotTo(HaveOccurred())
						Expect(disabled).To(BeTrue())
					})
				})
			})

			Context("when host tcp services are specified", func() {
				BeforeEach(func() {
					inputProperties = map[string]interface{}{
						".properties.host_tcp_services": []map[string]string{
							{
								"name":        "some-host-tcp-service",
								"ip_and_port": "169.254.0.1:2345",
							},
							{
								"name":        "some-other-host-tcp-service",
								"ip_and_port": "169.254.0.2:6789",
							},
						},
					}
				})

				It("adds the addresses to the manifest", func() {
					manifest, err := product.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(cellInstanceGroup, "silk-cni")
					Expect(err).NotTo(HaveOccurred())

					addresses, err := job.Property("host_tcp_services")
					Expect(err).NotTo(HaveOccurred())
					Expect(addresses).To(ConsistOf("169.254.0.1:2345", "169.254.0.2:6789"))
				})
			})
		})

		Context("when External is enabled", func() {
			BeforeEach(func() {
				inputProperties = map[string]interface{}{
					".properties.container_networking_interface_plugin": "external",
				}
			})

			It("configures the cni_config_dir", func() {
				manifest, err := product.RenderManifest(inputProperties)
				Expect(err).NotTo(HaveOccurred())

				job, err := manifest.FindInstanceGroupJob(cellInstanceGroup, "garden-cni")
				Expect(err).NotTo(HaveOccurred())

				cniConfigDir, err := job.Property("cni_config_dir")
				Expect(err).NotTo(HaveOccurred())
				Expect(cniConfigDir).To(Equal("/var/vcap/jobs/cni/config/cni"))

				cniPluginDir, err := job.Property("cni_plugin_dir")
				Expect(err).NotTo(HaveOccurred())
				Expect(cniPluginDir).To(Equal("/var/vcap/packages/cni/bin"))
			})
		})
	})

	Describe("DNS search domain", func() {
		var (
			inputProperties map[string]interface{}
			instanceGroup   string
		)

		BeforeEach(func() {
			if productName == "ert" {
				instanceGroup = "diego_cell"
			} else {
				instanceGroup = "compute"
			}
		})

		It("configures search_domains on the garden-cni job", func() {
			inputProperties = map[string]interface{}{
				".properties.cf_networking_search_domains": "some-search-domain,another-search-domain",
			}

			manifest, err := product.RenderManifest(inputProperties)
			Expect(err).NotTo(HaveOccurred())

			job, err := manifest.FindInstanceGroupJob(instanceGroup, "garden-cni")
			Expect(err).NotTo(HaveOccurred())

			searchDomains, err := job.Property("search_domains")
			Expect(err).NotTo(HaveOccurred())

			Expect(searchDomains).To(Equal([]interface{}{
				"some-search-domain",
				"another-search-domain",
			}))
		})

		It("configures search_domains on the garden-cni job", func() {
			manifest, err := product.RenderManifest(nil)
			Expect(err).NotTo(HaveOccurred())

			job, err := manifest.FindInstanceGroupJob(instanceGroup, "garden-cni")
			Expect(err).NotTo(HaveOccurred())

			searchDomains, err := job.Property("search_domains")
			Expect(err).NotTo(HaveOccurred())

			Expect(searchDomains).To(HaveLen(0))
		})
	})

	Describe("Service Discovery For Apps", func() {
		Describe("controller", func() {
			var instanceGroup string
			BeforeEach(func() {
				if productName == "ert" {
					instanceGroup = "diego_brain"
				} else {
					instanceGroup = "control"
				}
			})

			It("is deployed", func() {
				manifest, err := product.RenderManifest(nil)
				Expect(err).NotTo(HaveOccurred())

				_, err = manifest.FindInstanceGroupJob(instanceGroup, "service-discovery-controller")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("cell", func() {
			var instanceGroup string
			BeforeEach(func() {
				if productName == "ert" {
					instanceGroup = "diego_cell"
				} else {
					instanceGroup = "compute"
				}
			})

			It("co-locates the bosh-dns-adapter and bpm", func() {
				manifest, err := product.RenderManifest(nil)
				Expect(err).NotTo(HaveOccurred())

				_, err = manifest.FindInstanceGroupJob(instanceGroup, "bosh-dns-adapter")
				Expect(err).NotTo(HaveOccurred())

				_, err = manifest.FindInstanceGroupJob(instanceGroup, "bpm")
				Expect(err).NotTo(HaveOccurred())
			})

			It("emits internal routes", func() {
				manifest, err := product.RenderManifest(nil)
				Expect(err).NotTo(HaveOccurred())

				job, err := manifest.FindInstanceGroupJob(instanceGroup, "route_emitter")
				Expect(err).NotTo(HaveOccurred())

				enabled, err := job.Property("internal_routes/enabled")
				Expect(err).NotTo(HaveOccurred())

				Expect(enabled).To(BeTrue())
			})

			Context("when internal domain is empty", func() {
				It("defaults internal domain to apps.internal", func() {
					manifest, err := product.RenderManifest(nil)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(instanceGroup, "bosh-dns-adapter")
					Expect(err).NotTo(HaveOccurred())

					internalDomains, err := job.Property("internal_domains")
					Expect(err).NotTo(HaveOccurred())

					Expect(internalDomains).To(ConsistOf("apps.internal"))
				})
			})

			Context("when internal domains are configured", func() {
				var (
					inputProperties map[string]interface{}
				)

				It("sets internal domains to the provided internal domains", func() {
					inputProperties = map[string]interface{}{
						".properties.cf_networking_internal_domains": []map[string]interface{}{
							{
								"name": "some-internal-domain",
							},
							{
								"name": "some-other-internal-domain",
							},
						},
					}
					manifest, err := product.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(instanceGroup, "bosh-dns-adapter")
					Expect(err).NotTo(HaveOccurred())

					internalDomains, err := job.Property("internal_domains")
					Expect(err).NotTo(HaveOccurred())

					Expect(internalDomains).To(Equal([]interface{}{
						"some-internal-domain",
						"some-other-internal-domain",
					}))
				})
			})
		})

		Describe("api", func() {
			var instanceGroup string
			BeforeEach(func() {
				if productName == "ert" {
					instanceGroup = "cloud_controller"
				} else {
					instanceGroup = "control"
				}
			})

			Context("when internal domain is empty", func() {
				It("adds apps.internal to app domains", func() {
					manifest, err := product.RenderManifest(nil)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(instanceGroup, "cloud_controller_ng")
					Expect(err).NotTo(HaveOccurred())

					appDomains, err := job.Property("app_domains")
					Expect(err).NotTo(HaveOccurred())

					Expect(appDomains).To(Equal([]interface{}{
						"apps.example.com",
						[]interface{}{
							map[interface{}]interface{}{
								"internal": true,
								"name":     "apps.internal",
							},
						},
					}))
				})
			})

			Context("when internal domains are configured", func() {
				var (
					inputProperties map[string]interface{}
				)

				It("adds internal domains to app domains", func() {
					inputProperties = map[string]interface{}{
						".properties.cf_networking_internal_domains": []map[string]interface{}{
							{
								"name": "some-internal-domain",
							},
							{
								"name": "some-other-internal-domain",
							},
						},
					}
					manifest, err := product.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(instanceGroup, "cloud_controller_ng")
					Expect(err).NotTo(HaveOccurred())

					appDomains, err := job.Property("app_domains")
					Expect(err).NotTo(HaveOccurred())

					Expect(appDomains).To(Equal([]interface{}{
						"apps.example.com",
						[]interface{}{
							map[interface{}]interface{}{
								"name":     "some-internal-domain",
								"internal": true,
							},
							map[interface{}]interface{}{
								"name":     "some-other-internal-domain",
								"internal": true,
							},
						},
					}))
				})
			})
		})

		Describe("SSH proxy", func() {
			var instanceGroup string
			BeforeEach(func() {
				if productName == "srt" {
					instanceGroup = "control"
				} else {
					instanceGroup = "diego_brain"
				}
			})

			Context("when static IPs are set", func() {
				var inputProperties map[string]interface{}

				It("adds the static_ips", func() {
					p := ""
					if instanceGroup == "diego_brain" {
						p = ".diego_brain.static_ips"
					} else {
						p = ".control.static_ips"
					}

					inputProperties = map[string]interface{}{
						p: "0.0.0.0",
					}

					manifest, err := product.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					ips, err := manifest.Path(fmt.Sprintf("/instance_groups/name=%s/networks", instanceGroup))
					Expect(err).NotTo(HaveOccurred())

					key := ips.([]interface{})[0].(map[interface{}]interface{})
					Expect(key["static_ips"]).To(Equal([]interface{}{"0.0.0.0"}))
				})
			})
		})
	})

	Describe("Routing", func() {
		Describe("drain_timeout", func() {
			var (
				inputProperties     map[string]interface{}
				routerInstanceGroup string
			)

			BeforeEach(func() {
				routerInstanceGroup = "router"
			})

			Describe("when the property is set", func() {
				BeforeEach(func() {
					inputProperties = map[string]interface{}{
						".router.drain_timeout": 999,
					}
				})

				It("sets the drain_timeout", func() {
					manifest, err := product.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(routerInstanceGroup, "gorouter")
					Expect(err).NotTo(HaveOccurred())

					drainTimeout, err := job.Property("router/drain_timeout")
					Expect(err).NotTo(HaveOccurred())
					Expect(drainTimeout).To(Equal(999))
				})
			})

			Describe("when the property is not set", func() {
				BeforeEach(func() {
					inputProperties = map[string]interface{}{}
				})

				It("defaults to false", func() {
					manifest, err := product.RenderManifest(inputProperties)
					Expect(err).NotTo(HaveOccurred())

					job, err := manifest.FindInstanceGroupJob(routerInstanceGroup, "gorouter")
					Expect(err).NotTo(HaveOccurred())

					drainTimeout, err := job.Property("router/drain_timeout")
					Expect(err).NotTo(HaveOccurred())
					Expect(drainTimeout).To(Equal(900))
				})
			})
		})

		Context("logging_timestamp_format", func() {
			var (
				diegoCellInstanceGroup string
			)
			BeforeEach(func() {
				if productName == "srt" {
					diegoCellInstanceGroup = "compute"
				} else {
					diegoCellInstanceGroup = "diego_cell"
				}
			})

			When("logging_timestamp_format is set to deprecated", func() {
				It("is used in all jobs that have this property", func() {
					manifest := renderProductManifest(product, map[string]interface{}{
						".properties.logging_timestamp_format": "deprecated",
					})

					iptablesLogger := findManifestInstanceGroupJob(manifest, diegoCellInstanceGroup, "iptables-logger")
					loggingFormatTimestamp, err := iptablesLogger.Property("logging/format/timestamp")
					Expect(err).NotTo(HaveOccurred())
					Expect(loggingFormatTimestamp).To(Equal("deprecated"))

					silkDaemon := findManifestInstanceGroupJob(manifest, diegoCellInstanceGroup, "silk-daemon")
					loggingFormatTimestamp, err = silkDaemon.Property("logging/format/timestamp")
					Expect(err).NotTo(HaveOccurred())
					Expect(loggingFormatTimestamp).To(Equal("deprecated"))
				})
			})

			When("logging_format_timestamp is set to rfc3339", func() {
				It("is used in all jobs that have this property", func() {
					manifest := renderProductManifest(product, map[string]interface{}{
						".properties.logging_timestamp_format": "rfc3339",
					})

					iptablesLogger := findManifestInstanceGroupJob(manifest, diegoCellInstanceGroup, "iptables-logger")
					loggingFormatTimestamp, err := iptablesLogger.Property("logging/format/timestamp")
					Expect(err).NotTo(HaveOccurred())
					Expect(loggingFormatTimestamp).To(Equal("rfc3339"))

					silkDaemon := findManifestInstanceGroupJob(manifest, diegoCellInstanceGroup, "silk-daemon")
					loggingFormatTimestamp, err = silkDaemon.Property("logging/format/timestamp")
					Expect(err).NotTo(HaveOccurred())
					Expect(loggingFormatTimestamp).To(Equal("rfc3339"))
				})
			})
		})
		Describe("gorouter", func() {
			var (
				inputProperties     map[string]interface{}
				routerInstanceGroup string
			)

			BeforeEach(func() {
				routerInstanceGroup = "router"
				inputProperties = map[string]interface{}{
					".properties.router_headers_remove_if_specified": []map[string]interface{}{
						{
							"name": "header1",
						},
						{
							"name": "header2",
						},
					}}
			})

			It("sets the headers to be removed for http responses", func() {
				manifest, err := product.RenderManifest(inputProperties)
				Expect(err).NotTo(HaveOccurred())

				job, err := manifest.FindInstanceGroupJob(routerInstanceGroup, "gorouter")
				Expect(err).NotTo(HaveOccurred())

				removeHeaders, err := job.Property("router/http_rewrite/responses/remove_headers")
				Expect(err).NotTo(HaveOccurred())
				Expect(removeHeaders.([]interface{})[0].(map[interface{}]interface{})["name"]).To(Equal("header1"))
				Expect(removeHeaders.([]interface{})[1].(map[interface{}]interface{})["name"]).To(Equal("header2"))
			})
		})
	})
})
