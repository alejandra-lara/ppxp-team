package manifest_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/planitest"
)

var _ = Describe("Routing", func() {
	Describe("operator defaults", func() {
		It("configures the ha-proxy and router minimum TLS versions", func() {
			manifest, err := product.RenderManifest(nil)
			Expect(err).NotTo(HaveOccurred())

			haproxy, err := manifest.FindInstanceGroupJob("ha_proxy", "haproxy")
			Expect(err).NotTo(HaveOccurred())

			haproxyDisableTLS10, err := haproxy.Property("ha_proxy/disable_tls_10")
			Expect(err).NotTo(HaveOccurred())
			Expect(haproxyDisableTLS10).To(BeTrue())

			haproxyDisableTLS11, err := haproxy.Property("ha_proxy/disable_tls_11")
			Expect(err).NotTo(HaveOccurred())
			Expect(haproxyDisableTLS11).To(BeTrue())

			router, err := manifest.FindInstanceGroupJob("router", "gorouter")
			Expect(err).NotTo(HaveOccurred())

			routerMinTLSVersion, err := router.Property("router/min_tls_version")
			Expect(err).NotTo(HaveOccurred())
			Expect(routerMinTLSVersion).To(Equal("TLSv1.2"))
		})

		It("enables TLS to backends if a TLS route is registered", func() {
			manifest, err := product.RenderManifest(nil)
			Expect(err).NotTo(HaveOccurred())

			router, err := manifest.FindInstanceGroupJob("router", "gorouter")
			Expect(err).NotTo(HaveOccurred())

			tlsEnabled, err := router.Property("router/backends/enable_tls")
			Expect(err).NotTo(HaveOccurred())
			Expect(tlsEnabled).To(BeTrue())

			routerCACerts, err := router.Property("router/ca_certs")
			Expect(err).NotTo(HaveOccurred())
			Expect(routerCACerts).NotTo(BeEmpty())
		})

		Context("when the operator sets the minimum TLS version to 1.1", func() {
			var (
				manifest planitest.Manifest
				err      error
			)

			BeforeEach(func() {
				manifest, err = product.RenderManifest(map[string]interface{}{
					".properties.routing_minimum_tls_version": "tls_v1_1",
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("configures the ha-proxy and router minimum TLS versions", func() {
				haproxy, err := manifest.FindInstanceGroupJob("ha_proxy", "haproxy")
				Expect(err).NotTo(HaveOccurred())

				haproxyDisableTLS10, err := haproxy.Property("ha_proxy/disable_tls_10")
				Expect(err).NotTo(HaveOccurred())
				Expect(haproxyDisableTLS10).To(BeTrue())

				haproxyDisableTLS11, err := haproxy.Property("ha_proxy/disable_tls_11")
				Expect(err).NotTo(HaveOccurred())
				Expect(haproxyDisableTLS11).To(BeFalse())

				router, err := manifest.FindInstanceGroupJob("router", "gorouter")
				Expect(err).NotTo(HaveOccurred())

				routerMinTLSVersion, err := router.Property("router/min_tls_version")
				Expect(err).NotTo(HaveOccurred())
				Expect(routerMinTLSVersion).To(Equal("TLSv1.1"))
			})
		})
	})

	Describe("TLS termination", func() {
		It("secures traffic between the infrastructure load balancer and HAProxy / Gorouter", func() {
			manifest, err := product.RenderManifest(nil)
			Expect(err).NotTo(HaveOccurred())

			haproxy, err := manifest.FindInstanceGroupJob("ha_proxy", "haproxy")
			Expect(err).NotTo(HaveOccurred())

			haproxySSLPEM, err := haproxy.Property("ha_proxy/ssl_pem")
			Expect(err).NotTo(HaveOccurred())
			Expect(haproxySSLPEM).NotTo(BeEmpty())

			router, err := manifest.FindInstanceGroupJob("router", "gorouter")
			Expect(err).NotTo(HaveOccurred())

			routerEnableSSL, err := router.Property("router/enable_ssl")
			Expect(err).NotTo(HaveOccurred())
			Expect(routerEnableSSL).To(BeTrue())

			routerTLSPEM, err := router.Property("router/tls_pem")
			Expect(err).NotTo(HaveOccurred())
			Expect(routerTLSPEM).NotTo(BeEmpty())
		})
	})

	Describe("IP Logging", func() {
		Context("when the operator chooses to log client Ips", func() {
			It("does not disable ip logging or x-forwarded-for logging", func() {
				manifest, err := product.RenderManifest(map[string]interface{}{
					".properties.routing_log_client_ips": "log_client_ips",
				})
				Expect(err).NotTo(HaveOccurred())

				router, err := manifest.FindInstanceGroupJob("router", "gorouter")
				Expect(err).NotTo(HaveOccurred())

				disableLogForwardedFor, err := router.Property("router/disable_log_forwarded_for")
				Expect(err).NotTo(HaveOccurred())
				Expect(disableLogForwardedFor).To(BeFalse())

				disableLogSourceIPs, err := router.Property("router/disable_log_source_ips")
				Expect(err).NotTo(HaveOccurred())
				Expect(disableLogSourceIPs).To(BeFalse())
			})
		})
		Context("when the operator chooses `Disable logging of X-Forwarded-For header only`", func() {
			It("only disables x-forwarded-for logging but not source ip logging", func() {
				manifest, err := product.RenderManifest(map[string]interface{}{
					".properties.routing_log_client_ips": "disable_x_forwarded_for",
				})
				Expect(err).NotTo(HaveOccurred())

				router, err := manifest.FindInstanceGroupJob("router", "gorouter")
				Expect(err).NotTo(HaveOccurred())

				disableLogForwardedFor, err := router.Property("router/disable_log_forwarded_for")
				Expect(err).NotTo(HaveOccurred())
				Expect(disableLogForwardedFor).To(BeTrue())

				disableLogSourceIPs, err := router.Property("router/disable_log_source_ips")
				Expect(err).NotTo(HaveOccurred())
				Expect(disableLogSourceIPs).To(BeFalse())
			})
		})
		Context("when the operator chooses `Disable logging of both source IP and X-Forwarded-For header`", func() {
			It("disbales both source ip logging and x-forwarded-for logging", func() {
				manifest, err := product.RenderManifest(map[string]interface{}{
					".properties.routing_log_client_ips": "disable_all_log_client_ips",
				})
				Expect(err).NotTo(HaveOccurred())

				router, err := manifest.FindInstanceGroupJob("router", "gorouter")
				Expect(err).NotTo(HaveOccurred())

				disableLogForwardedFor, err := router.Property("router/disable_log_forwarded_for")
				Expect(err).NotTo(HaveOccurred())
				Expect(disableLogForwardedFor).To(BeTrue())

				disableLogSourceIPs, err := router.Property("router/disable_log_source_ips")
				Expect(err).NotTo(HaveOccurred())
				Expect(disableLogSourceIPs).To(BeTrue())
			})
		})
	})

	// TODO: stop skipping once ops-manifest supports testing for credentials
	XDescribe("Gorouter provides client certs in request to Diego cells", func() {
		It("creates a backend cert_chain and private_key", func() {
			manifest, err := product.RenderManifest(map[string]interface{}{})
			Expect(err).NotTo(HaveOccurred())

			router, err := manifest.FindInstanceGroupJob("router", "gorouter")
			Expect(err).NotTo(HaveOccurred())

			certChain, err := router.Property("router/backends/cert_chain")
			Expect(err).NotTo(HaveOccurred())
			Expect(certChain).NotTo(BeNil())

			privateKey, err := router.Property("router/backends/private_key")
			Expect(err).NotTo(HaveOccurred())
			Expect(privateKey).NotTo(BeNil())
		})
	})

	Describe("Router Client Cert Validation", func() {
		Context("when it does not request client certificates", func() {
			It("sets the validation type to none", func() {
				manifest, err := product.RenderManifest(map[string]interface{}{
					".properties.router_client_cert_validation": "none",
				})
				Expect(err).NotTo(HaveOccurred())

				router, err := manifest.FindInstanceGroupJob("router", "gorouter")
				Expect(err).NotTo(HaveOccurred())
				Expect(router.Property("router/client_cert_validation")).To(ContainSubstring("none"))
			})
		})

		Context("when it requests but does not require client certificates", func() {
			It("sets the validation type to request", func() {
				manifest, err := product.RenderManifest(map[string]interface{}{})
				Expect(err).NotTo(HaveOccurred())

				router, err := manifest.FindInstanceGroupJob("router", "gorouter")
				Expect(err).NotTo(HaveOccurred())
				Expect(router.Property("router/client_cert_validation")).To(ContainSubstring("request"))
			})
		})

		Context("when it requires client certificates", func() {
			It("sets the validation type to require", func() {
				manifest, err := product.RenderManifest(map[string]interface{}{
					".properties.router_client_cert_validation": "require",
				})
				Expect(err).NotTo(HaveOccurred())

				router, err := manifest.FindInstanceGroupJob("router", "gorouter")
				Expect(err).NotTo(HaveOccurred())
				Expect(router.Property("router/client_cert_validation")).To(ContainSubstring("require"))
			})
		})
	})

	Describe("TLS termination", func() {
		Context("when TLS is terminated for the first time at infrastructure load balancer", func() {
			It("configures the router and proxy", func() {
				manifest, err := product.RenderManifest(map[string]interface{}{})
				Expect(err).NotTo(HaveOccurred())

				haproxy, err := manifest.FindInstanceGroupJob("ha_proxy", "haproxy")
				Expect(err).NotTo(HaveOccurred())
				Expect(haproxy.Property("ha_proxy")).ShouldNot(HaveKey("client_ca_file"))
				Expect(haproxy.Property("ha_proxy/client_cert")).To(BeFalse())

				router, err := manifest.FindInstanceGroupJob("router", "gorouter")
				Expect(err).NotTo(HaveOccurred())
				Expect(router.Property("router/forwarded_client_cert")).To(ContainSubstring("always_forward"))
			})
		})

		Context("when TLS is terminated for the first time at ha proxy", func() {
			Context("when ha proxy client cert validation is set to none", func() {
				It("configures ha proxy and router", func() {
					manifest, err := product.RenderManifest(map[string]interface{}{
						".properties.routing_tls_termination": "ha_proxy",
					})
					Expect(err).NotTo(HaveOccurred())

					haproxy, err := manifest.FindInstanceGroupJob("ha_proxy", "haproxy")
					Expect(err).NotTo(HaveOccurred())
					Expect(haproxy.Property("ha_proxy")).ShouldNot(HaveKey("client_ca_file"))
					Expect(haproxy.Property("ha_proxy/client_cert")).To(BeFalse())

					router, err := manifest.FindInstanceGroupJob("router", "gorouter")
					Expect(err).NotTo(HaveOccurred())
					Expect(router.Property("router/forwarded_client_cert")).To(ContainSubstring("forward"))
				})
			})

			Context("when ha proxy client cert validation is set to request ", func() {
				It("configures ha proxy and router", func() {
					manifest, err := product.RenderManifest(map[string]interface{}{
						".properties.routing_tls_termination":        "ha_proxy",
						".properties.haproxy_client_cert_validation": "request",
					})
					Expect(err).NotTo(HaveOccurred())

					haproxy, err := manifest.FindInstanceGroupJob("ha_proxy", "haproxy")
					Expect(err).NotTo(HaveOccurred())
					Expect(haproxy.Property("ha_proxy")).ShouldNot(HaveKey("client_ca_file"))
					Expect(haproxy.Property("ha_proxy/client_cert")).To(BeTrue())

					router, err := manifest.FindInstanceGroupJob("router", "gorouter")
					Expect(err).NotTo(HaveOccurred())
					Expect(router.Property("router/forwarded_client_cert")).To(ContainSubstring("forward"))
				})
			})
		})

		Context("when TLS is terminated for the first time at the router", func() {
			It("configures the router and proxy", func() {
				manifest, err := product.RenderManifest(map[string]interface{}{
					".properties.routing_tls_termination": "router",
				})
				Expect(err).NotTo(HaveOccurred())

				haproxy, err := manifest.FindInstanceGroupJob("ha_proxy", "haproxy")
				Expect(err).NotTo(HaveOccurred())
				Expect(haproxy.Property("ha_proxy")).ShouldNot(HaveKey("client_ca_file"))
				Expect(haproxy.Property("ha_proxy/client_cert")).To(BeFalse())

				router, err := manifest.FindInstanceGroupJob("router", "gorouter")
				Expect(err).NotTo(HaveOccurred())
				Expect(router.Property("router/forwarded_client_cert")).To(ContainSubstring("sanitize_set"))
			})
		})
	})

	Describe("idle timeouts", func() {

		It("sets a default timeout", func() {
			manifest, err := product.RenderManifest(nil)
			Expect(err).NotTo(HaveOccurred())

			haproxy, err := manifest.FindInstanceGroupJob("ha_proxy", "haproxy")
			Expect(err).NotTo(HaveOccurred())
			haproxyTimeout, err := haproxy.Property("ha_proxy/keepalive_timeout")
			Expect(err).NotTo(HaveOccurred())
			Expect(haproxyTimeout).To(Equal(900))

			router, err := manifest.FindInstanceGroupJob("router", "gorouter")
			Expect(err).NotTo(HaveOccurred())
			routerTimeout, err := router.Property("router/frontend_idle_timeout")
			Expect(err).NotTo(HaveOccurred())
			Expect(routerTimeout).To(Equal(900))
		})

		Context("when the operator specifies an idle timeout for IaaS compatibility", func() {

			It("is applied", func() {
				manifest, err := product.RenderManifest(map[string]interface{}{
					".router.frontend_idle_timeout": 300,
				})
				Expect(err).NotTo(HaveOccurred())

				haproxy, err := manifest.FindInstanceGroupJob("ha_proxy", "haproxy")
				Expect(err).NotTo(HaveOccurred())
				haproxyTimeout, err := haproxy.Property("ha_proxy/keepalive_timeout")
				Expect(err).NotTo(HaveOccurred())
				Expect(haproxyTimeout).To(Equal(300))

				router, err := manifest.FindInstanceGroupJob("router", "gorouter")
				Expect(err).NotTo(HaveOccurred())
				routerTimeout, err := router.Property("router/frontend_idle_timeout")
				Expect(err).NotTo(HaveOccurred())
				Expect(routerTimeout).To(Equal(300))
			})

		})

	})

	Describe("BPM", func() {
		var routingJobs []Job

		BeforeEach(func() {
			if productName == "srt" {
				routingJobs = []Job{
					{
						Name:          "route_registrar",
						InstanceGroup: "control",
					},
					{
						Name:          "route_registrar",
						InstanceGroup: "database",
					},
					{
						Name:          "route_registrar",
						InstanceGroup: "blobstore",
					},
					{
						Name:          "tcp_router",
						InstanceGroup: "tcp_router",
					},
					{
						Name:          "routing-api",
						InstanceGroup: "control",
					},
					{
						Name:          "gorouter",
						InstanceGroup: "router",
					},
					{
						Name:          "bbr-routingdb",
						InstanceGroup: "backup-prepare",
					},
				}
			} else {
				routingJobs = []Job{
					{
						Name:          "tcp_router",
						InstanceGroup: "tcp_router",
					},
					{
						Name:          "routing-api",
						InstanceGroup: "cloud_controller",
					},
					{
						Name:          "route_registrar",
						InstanceGroup: "cloud_controller",
					},
					{
						Name:          "route_registrar",
						InstanceGroup: "loggregator_trafficcontroller",
					},
					{
						Name:          "route_registrar",
						InstanceGroup: "nfs_server",
					},
					{
						Name:          "route_registrar",
						InstanceGroup: "mysql_proxy",
					},
					{
						Name:          "route_registrar",
						InstanceGroup: "diego_database",
					},
					{
						Name:          "route_registrar",
						InstanceGroup: "doppler",
					},
					{
						Name:          "route_registrar",
						InstanceGroup: "uaa",
					},
					{
						Name:          "gorouter",
						InstanceGroup: "router",
					},
					{
						Name:          "bbr-routingdb",
						InstanceGroup: "backup-prepare",
					},
				}
			}
		})

		It("co-locates the BPM job with all routing jobs", func() {
			manifest, err := product.RenderManifest(nil)
			Expect(err).NotTo(HaveOccurred())

			for _, job := range routingJobs {
				_, err = manifest.FindInstanceGroupJob(job.InstanceGroup, "bpm")
				Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Expected to find `bpm` job on instance group `%s`", job.InstanceGroup))
			}
		})
	})
})
