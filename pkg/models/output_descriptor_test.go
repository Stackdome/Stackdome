package models

import (
	"testing"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func TestPublicURLOutputs(t *testing.T) {
	for _, tc := range []struct {
		name    string
		config  PublicEndpointConfig
		fqdn    string
		public  bool
		wantURL string
	}{
		{"shared TLS", PublicEndpointConfig{SharedCompute: true, PlatformTLSEnabled: true}, "web.example.com", true, "https://web.example.com"},
		{"shared HTTP", PublicEndpointConfig{SharedCompute: true}, "web.example.com", true, "http://web.example.com"},
		{"BYOC TLS", PublicEndpointConfig{}, "web.example.com", true, "https://web.example.com"},
		{"local HTTP", PublicEndpointConfig{}, "web.local", true, "http://web.local"},
		{"private port", PublicEndpointConfig{}, "web.example.com", false, ""},
		{"unassigned hostname", PublicEndpointConfig{}, "", true, ""},
	} {
		for _, layout := range []string{"single", "multiple"} {
			t.Run(tc.name+"/"+layout, func(t *testing.T) {
				g := gomega.NewWithT(t)
				r := &StackResource{Name: "web", Namespace: "app", Ports: Ports{
					{Name: "http", Number: 8080, Protocol: PortProtocolHTTP, ExposedToPublic: tc.public, ExposedFqdn: tc.fqdn},
				}}
				suffix := ""
				if layout == "multiple" {
					r.Ports = append(r.Ports, Port{Name: "db", Number: 5432, Protocol: PortProtocolTCP})
					suffix = ".http"
				}
				outputs := r.ToOutputMap(tc.config)
				g.Expect(outputs[OutputNameURL+suffix]).To(gomega.Equal("http://web.app.svc:8080"))
				if tc.wantURL == "" {
					g.Expect(outputs).NotTo(gomega.HaveKey(OutputNamePublicURL + suffix))
					g.Expect(outputs).NotTo(gomega.HaveKey(OutputNamePublicHost + suffix))
				} else {
					g.Expect(outputs[OutputNamePublicURL+suffix]).To(gomega.Equal(tc.wantURL))
					g.Expect(outputs[OutputNamePublicHost+suffix]).To(gomega.Equal(tc.fqdn))
				}
				if layout == "multiple" {
					g.Expect(outputs["url.db"]).To(gomega.Equal("web.app.svc:5432"))
					g.Expect(outputs).NotTo(gomega.HaveKey("public_url.db"))
				}
			})
		}
	}
}

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
