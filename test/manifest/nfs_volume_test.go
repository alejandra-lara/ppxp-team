package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NFS volume service", func() {
	var instanceGroup string

	Context("when the NFS V3 driver is enabled without LDAP configuration", func() {
		It("disables LDAP on the nfsbrokerpush job", func() {
			if productName == "srt" {
				instanceGroup = "control"
			} else {
				instanceGroup = "clock_global"
			}

			manifest, err := product.RenderManifest(nil)
			Expect(err).NotTo(HaveOccurred())

			nfsBrokerPush, err := manifest.FindInstanceGroupJob(instanceGroup, "nfsbrokerpush")
			Expect(err).NotTo(HaveOccurred())

			ldapEnabled, err := nfsBrokerPush.Property("nfsbrokerpush/ldap_enabled")
			Expect(err).NotTo(HaveOccurred())
			Expect(ldapEnabled).To(BeFalse())
		})

		It("does not configure LDAP on the nfsv3driver job", func() {
			if productName == "srt" {
				instanceGroup = "compute"
			} else {
				instanceGroup = "diego_cell"
			}

			manifest, err := product.RenderManifest(nil)
			Expect(err).NotTo(HaveOccurred())

			nfsV3Driver, err := manifest.FindInstanceGroupJob(instanceGroup, "nfsv3driver")
			Expect(err).NotTo(HaveOccurred())

			nfsV3DriverProperties, err := nfsV3Driver.Property("nfsv3driver")
			Expect(err).NotTo(HaveOccurred())

			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("disable", BeFalse()))
			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("ldap_svc_user", BeNil()))
			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("ldap_svc_password", BeNil()))
			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("ldap_host", BeNil()))
			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("ldap_port", BeNil()))
			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("ldap_user_fqdn", BeNil()))
			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("ldap_ca_cert", BeNil()))
			Expect(nfsV3DriverProperties).NotTo(HaveKey("allowed-in-source"))
		})
	})

	Context("when the NFS V3 driver is enabled with LDAP configuration", func() {
		var ldapConfiguration map[string]interface{}

		BeforeEach(func() {
			ldapConfiguration = map[string]interface{}{
				".properties.nfs_volume_driver.enable.ldap_service_account_user": "service-account-user",
				".properties.nfs_volume_driver.enable.ldap_service_account_password": map[string]string{
					"secret": "service-account-password",
				},
				".properties.nfs_volume_driver.enable.ldap_server_host": "ldap-host",
				".properties.nfs_volume_driver.enable.ldap_server_port": 12345,
				".properties.nfs_volume_driver.enable.ldap_user_fqdn":   "ldap-user-search-base",
			}
		})

		It("enables LDAP on the nfsbrokerpush job", func() {
			if productName == "srt" {
				instanceGroup = "control"
			} else {
				instanceGroup = "clock_global"
			}

			manifest, err := product.RenderManifest(ldapConfiguration)
			Expect(err).NotTo(HaveOccurred())

			nfsBrokerPush, err := manifest.FindInstanceGroupJob(instanceGroup, "nfsbrokerpush")
			Expect(err).NotTo(HaveOccurred())

			ldapEnabled, err := nfsBrokerPush.Property("nfsbrokerpush/ldap_enabled")
			Expect(err).NotTo(HaveOccurred())
			Expect(ldapEnabled).To(BeTrue())
		})

		It("configures LDAP on the nfsv3driver job", func() {
			if productName == "srt" {
				instanceGroup = "compute"
			} else {
				instanceGroup = "diego_cell"
			}

			manifest, err := product.RenderManifest(ldapConfiguration)
			Expect(err).NotTo(HaveOccurred())

			nfsV3Driver, err := manifest.FindInstanceGroupJob(instanceGroup, "nfsv3driver")
			Expect(err).NotTo(HaveOccurred())

			nfsV3DriverProperties, err := nfsV3Driver.Property("nfsv3driver")
			Expect(err).NotTo(HaveOccurred())

			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("disable", BeFalse()))
			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("ldap_svc_user", "service-account-user"))
			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("ldap_svc_password", Not(BeEmpty())))
			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("ldap_host", "ldap-host"))
			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("ldap_port", 12345))
			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("ldap_user_fqdn", "ldap-user-search-base"))
			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("ldap_ca_cert", BeNil()))
			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("allowed-in-source", ""))
		})

		Context("when the LDAP CA certificate is configured", func() {
			It("configures LDAP on the nfsv3driver job", func() {
				if productName == "srt" {
					instanceGroup = "compute"
				} else {
					instanceGroup = "diego_cell"
				}

				ldapConfiguration[".properties.nfs_volume_driver.enable.ldap_ca_cert"] = "ldap-ca-cert"

				manifest, err := product.RenderManifest(ldapConfiguration)
				Expect(err).NotTo(HaveOccurred())

				nfsV3Driver, err := manifest.FindInstanceGroupJob(instanceGroup, "nfsv3driver")
				Expect(err).NotTo(HaveOccurred())

				nfsV3DriverProperties, err := nfsV3Driver.Property("nfsv3driver")
				Expect(err).NotTo(HaveOccurred())

				Expect(nfsV3DriverProperties).To(HaveKeyWithValue("ldap_ca_cert", "ldap-ca-cert"))
			})
		})
	})

	Context("when the NFS V3 driver is disabled", func() {
		It("disables the nfsv3driver job", func() {
			if productName == "srt" {
				instanceGroup = "compute"
			} else {
				instanceGroup = "diego_cell"
			}

			manifest, err := product.RenderManifest(map[string]interface{}{
				".properties.nfs_volume_driver": "disable",
			})
			Expect(err).NotTo(HaveOccurred())

			nfsV3Driver, err := manifest.FindInstanceGroupJob(instanceGroup, "nfsv3driver")
			Expect(err).NotTo(HaveOccurred())

			nfsV3DriverProperties, err := nfsV3Driver.Property("nfsv3driver")
			Expect(err).NotTo(HaveOccurred())

			Expect(nfsV3DriverProperties).To(HaveKeyWithValue("disable", BeTrue()))
		})
	})
})
