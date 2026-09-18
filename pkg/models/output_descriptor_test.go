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
	buildOutputs := func(multiPort bool, config PublicEndpointConfig, fqdn string, exposed bool) map[string]string {
		ports := Ports{
			{Name: "http", Number: 8080, Protocol: PortProtocolHTTP, ExposedToPublic: exposed, ExposedFqdn: fqdn},
		}
		if multiPort {
			ports = append(ports, Port{Name: "metrics", Number: 9090, Protocol: PortProtocolTCP})
		}
		return (&StackResource{Name: "web", Namespace: "app", Ports: ports}).ToOutputMap(config)
	}

	ginkgo.DescribeTable("derives the scheme from the public endpoint TLS policy",
		func(multiPort bool, config PublicEndpointConfig, fqdn, wantURL string) {
			suffix := ""
			if multiPort {
				suffix = ".http"
			}
			outputs := buildOutputs(multiPort, config, fqdn, true)
			gomega.Expect(outputs[OutputNamePublicURL+suffix]).To(gomega.Equal(wantURL))
			gomega.Expect(outputs[OutputNamePublicHost+suffix]).To(gomega.Equal(fqdn))
		},
		ginkgo.Entry("HTTPS single port", false, PublicEndpointConfig{}, "web.example.com", "https://web.example.com"),
		ginkgo.Entry("HTTP single port", false, PublicEndpointConfig{}, "web.local", "http://web.local"),
		ginkgo.Entry("HTTPS multi port", true, PublicEndpointConfig{}, "web.example.com", "https://web.example.com"),
		ginkgo.Entry("HTTP multi port", true, PublicEndpointConfig{}, "web.local", "http://web.local"),
		ginkgo.Entry("shared compute without platform TLS", false, PublicEndpointConfig{SharedCompute: true}, "web.example.com", "http://web.example.com"),
		ginkgo.Entry("shared compute with platform TLS", false, PublicEndpointConfig{SharedCompute: true, PlatformTLSEnabled: true}, "web.example.com", "https://web.example.com"),
		ginkgo.Entry("shared compute without platform TLS (multi port)", true, PublicEndpointConfig{SharedCompute: true}, "web.example.com", "http://web.example.com"),
		ginkgo.Entry("shared compute with platform TLS (multi port)", true, PublicEndpointConfig{SharedCompute: true, PlatformTLSEnabled: true}, "web.example.com", "https://web.example.com"),
	)

	ginkgo.It("keeps internal service URLs on the service host regardless of endpoint TLS", func() {
		outputs := buildOutputs(true, PublicEndpointConfig{SharedCompute: true, PlatformTLSEnabled: true}, "web.example.com", true)
		gomega.Expect(outputs["url.http"]).To(gomega.Equal("http://web.app.svc:8080"))
		gomega.Expect(outputs["url.metrics"]).To(gomega.Equal("web.app.svc:9090"))
	})

	ginkgo.It("omits public outputs for private or unassigned ports", func() {
		for _, outputs := range []map[string]string{
			buildOutputs(false, PublicEndpointConfig{}, "web.example.com", false),
			buildOutputs(false, PublicEndpointConfig{}, "", true),
		} {
			gomega.Expect(outputs).NotTo(gomega.HaveKey(OutputNamePublicURL))
			gomega.Expect(outputs).NotTo(gomega.HaveKey(OutputNamePublicHost))
		}
	})
})
