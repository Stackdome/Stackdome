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

var _ = ginkgo.Describe("StackResource public URL scheme", func() {
	sharedTLS := PublicEndpointConfig{SharedCompute: true, PlatformTLSEnabled: true}
	sharedHTTP := PublicEndpointConfig{SharedCompute: true}
	byoc := PublicEndpointConfig{}

	type schemeCase struct {
		name    string
		config  PublicEndpointConfig
		fqdn    string
		wantURL string
	}

	ginkgo.DescribeTable("derives the scheme from the public endpoint TLS policy",
		func(tc schemeCase) {
			r := &StackResource{Name: "web", Namespace: "app", Ports: Ports{
				{Name: "http", Number: 8080, Protocol: PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: tc.fqdn},
			}}
			m := r.ToOutputMap(tc.config)
			gomega.Expect(m[OutputNamePublicHost]).To(gomega.Equal(tc.fqdn))
			gomega.Expect(m[OutputNamePublicURL]).To(gomega.Equal(tc.wantURL))
			gomega.Expect(m[OutputNameURL]).To(gomega.Equal("http://web.app.svc:8080"))
		},
		ginkgo.Entry("shared compute with platform TLS", schemeCase{"shared tls", sharedTLS, "api.app.stackdome.com", "https://api.app.stackdome.com"}),
		ginkgo.Entry("shared compute without platform TLS", schemeCase{"shared http", sharedHTTP, "api.app.stackdome.com", "http://api.app.stackdome.com"}),
		ginkgo.Entry("BYOC real domain", schemeCase{"byoc tls", byoc, "api.example.com", "https://api.example.com"}),
		ginkgo.Entry("BYOC local hostname", schemeCase{"byoc local", byoc, "api.local", "http://api.local"}),
		ginkgo.Entry("BYOC nip.io hostname", schemeCase{"byoc nip", byoc, "api.127-0-0-1.nip.io", "http://api.127-0-0-1.nip.io"}),
	)

	ginkgo.It("applies the scheme to per-port keys on a multi-port resource", func() {
		r := &StackResource{Name: "web", Namespace: "app", Ports: Ports{
			{Name: "http", Number: 8080, Protocol: PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: "api.example.com"},
			{Name: "metrics", Number: 9090, Protocol: PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: "metrics.local"},
		}}
		m := r.ToOutputMap(byoc)
		gomega.Expect(m[OutputNamePublicURL+".http"]).To(gomega.Equal("https://api.example.com"))
		gomega.Expect(m[OutputNamePublicURL+".metrics"]).To(gomega.Equal("http://metrics.local"))
		gomega.Expect(m[OutputNameURL+".http"]).To(gomega.Equal("http://web.app.svc:8080"))
		gomega.Expect(m[OutputNameURL+".metrics"]).To(gomega.Equal("http://web.app.svc:9090"))
	})

	ginkgo.It("omits public outputs for private ports and unassigned hostnames", func() {
		r := &StackResource{Name: "web", Namespace: "app", Ports: Ports{
			{Name: "private", Number: 8080, Protocol: PortProtocolHTTP, ExposedToPublic: false, ExposedFqdn: "api.example.com"},
			{Name: "unassigned", Number: 9090, Protocol: PortProtocolHTTP, ExposedToPublic: true, ExposedFqdn: ""},
		}}
		m := r.ToOutputMap(sharedTLS)
		gomega.Expect(m).NotTo(gomega.HaveKey(OutputNamePublicURL + ".private"))
		gomega.Expect(m).NotTo(gomega.HaveKey(OutputNamePublicURL + ".unassigned"))
		gomega.Expect(m).NotTo(gomega.HaveKey(OutputNamePublicHost + ".private"))
		gomega.Expect(m).NotTo(gomega.HaveKey(OutputNamePublicHost + ".unassigned"))
	})
})
