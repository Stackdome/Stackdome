package models

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("StackResource output naming", func() {
	newResource := func(ports ...Port) *StackResource {
		return &StackResource{Name: "mysql", Ports: ports}
	}
	names := func(r *StackResource) []string {
		out := []string{}
		for _, d := range StackResourceOutputDescriptors(r) {
			out = append(out, d.Name)
		}
		return out
	}

	ginkgo.It("emits only host for a zero-port resource", func() {
		gomega.Expect(names(newResource())).To(gomega.Equal([]string{OutputNameHost}))
	})

	ginkgo.It("drops the port suffix for a single-port resource", func() {
		r := newResource(Port{Name: "3306", Number: 3306, Protocol: PortProtocolTCP})
		gomega.Expect(names(r)).To(gomega.Equal([]string{
			OutputNameHost, OutputNamePort, OutputNameURL,
		}))
	})

	ginkgo.It("adds public_host/public_url (unsuffixed) for a single public port", func() {
		r := newResource(Port{Name: "80", Number: 80, Protocol: PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: "web.example.com"})
		gomega.Expect(names(r)).To(gomega.Equal([]string{
			OutputNameHost, OutputNamePort, OutputNameURL,
			OutputNamePublicHost, OutputNamePublicURL,
		}))
	})

	ginkgo.It("suffixes every per-port key with the port name for a multi-port resource", func() {
		r := newResource(
			Port{Name: "3306", Number: 3306, Protocol: PortProtocolTCP},
			Port{Name: "metrics", Number: 9090, Protocol: PortProtocolHTTP},
		)
		gomega.Expect(names(r)).To(gomega.Equal([]string{
			OutputNameHost,
			"port.3306", "url.3306",
			"port.metrics", "url.metrics",
		}))
	})

	ginkgo.It("keeps ToOutputMap keys identical to the descriptor names (no drift)", func() {
		cases := []*StackResource{
			newResource(),
			newResource(Port{Name: "3306", Number: 3306, Protocol: PortProtocolTCP}),
			newResource(Port{Name: "80", Number: 80, Protocol: PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: "web.example.com"}),
			newResource(
				Port{Name: "3306", Number: 3306, Protocol: PortProtocolTCP},
				Port{Name: "80", Number: 80, Protocol: PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: "web.example.com"},
			),
		}
		for _, r := range cases {
			descNames := map[string]bool{}
			for _, d := range StackResourceOutputDescriptors(r) {
				descNames[d.Name] = true
			}
			for k := range r.ToOutputMap(PublicEndpointConfig{}) {
				gomega.Expect(descNames).To(gomega.HaveKey(k), "ToOutputMap key %q missing from descriptors", k)
			}
			gomega.Expect(len(r.ToOutputMap(PublicEndpointConfig{}))).To(gomega.Equal(len(descNames)))
		}
	})

	ginkgo.It("emits multi-port values under the suffixed keys", func() {
		r := newResource(
			Port{Name: "3306", Number: 3306, Protocol: PortProtocolTCP},
			Port{Name: "80", Number: 80, Protocol: PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: "web.example.com"},
		)
		m := r.ToOutputMap(PublicEndpointConfig{})
		gomega.Expect(m["url.3306"]).To(gomega.Equal("mysql:3306"))
		gomega.Expect(m["url.80"]).To(gomega.Equal("http://mysql:80"))
		gomega.Expect(m[OutputNamePublicURL+".80"]).To(gomega.Equal("https://web.example.com"))
	})
})

var _ = ginkgo.Describe("StackResource public URL outputs", func() {
	ginkgo.DescribeTable("uses the public endpoint TLS scheme",
		func(config PublicEndpointConfig, fqdn, wantScheme string) {
			for _, tc := range []struct {
				name   string
				ports  Ports
				suffix string
			}{
				{
					name:  "single port",
					ports: Ports{{Name: "http", Number: 8080, Protocol: PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: fqdn}},
				},
				{
					name: "multi port",
					ports: Ports{
						{Name: "http", Number: 8080, Protocol: PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: fqdn},
						{Name: "metrics", Number: 9090, Protocol: PortProtocolTCP},
					},
					suffix: ".http",
				},
			} {
				ginkgo.By(tc.name)
				r := &StackResource{Name: "web", Namespace: "app", Ports: tc.ports}
				m := r.ToOutputMap(config)
				gomega.Expect(m[OutputNamePublicURL+tc.suffix]).To(gomega.Equal(wantScheme + "://" + fqdn))
				gomega.Expect(m[OutputNamePublicHost+tc.suffix]).To(gomega.Equal(fqdn))
				gomega.Expect(m[OutputNameURL+tc.suffix]).To(gomega.Equal("http://web.app.svc:8080"))
			}
		},
		ginkgo.Entry("HTTP endpoint stays http", PublicEndpointConfig{}, "api.127-0-0-1.nip.io", "http"),
		ginkgo.Entry("HTTPS endpoint uses https", PublicEndpointConfig{}, "api.customer.example.com", "https"),
		ginkgo.Entry("shared compute without platform TLS stays http", PublicEndpointConfig{SharedCompute: true}, "api-1234abcd.cloud.stackdome.com", "http"),
		ginkgo.Entry("shared compute with platform TLS uses https", PublicEndpointConfig{SharedCompute: true, PlatformTLSEnabled: true}, "api-1234abcd.cloud.stackdome.com", "https"),
	)

	ginkgo.It("keeps internal service URLs HTTP-only and omits private public keys", func() {
		r := &StackResource{
			Name: "mysql", Namespace: "app",
			Ports: Ports{
				{Name: "http", Number: 8080, Protocol: PortProtocolHTTP},
				{Name: "tcp", Number: 3306, Protocol: PortProtocolTCP, ExposedToPublic: true, ExposedFqdn: "db.example.com"},
			},
		}
		m := r.ToOutputMap(PublicEndpointConfig{})
		gomega.Expect(m["url.http"]).To(gomega.Equal("http://mysql.app.svc:8080"))
		gomega.Expect(m["url.tcp"]).To(gomega.Equal("mysql.app.svc:3306"))
		gomega.Expect(m).NotTo(gomega.HaveKey(OutputNamePublicURL + ".http"))
		gomega.Expect(m[OutputNamePublicURL+".tcp"]).To(gomega.Equal("https://db.example.com"))
	})
})
