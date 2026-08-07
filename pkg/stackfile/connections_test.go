package stackfile

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/Stackdome/stackdome/pkg/api/openapi"
)

var _ = Describe("superuser addon connections", func() {
	// One addon, two consumers: the app on owner credentials, the
	// migration Job on superuser credentials.
	var stack openapi.Stack

	BeforeEach(func() {
		sf, err := Load(loadFixtureBytes("superuser_migration.yaml"))
		Expect(err).NotTo(HaveOccurred())
		stack, err = sf.ToStack()
		Expect(err).NotTo(HaveOccurred())
	})

	addonConnTo := func(target string) openapi.StackConnection {
		for _, c := range stack.Spec.Connections {
			if c.From.Type == "addon/"+PostgresAddonType && c.To.GetName() == target {
				return c
			}
		}
		Fail("no addon connection to " + target)
		return openapi.StackConnection{}
	}

	It("gives the app owner credentials scoped to the database", func() {
		conn := addonConnTo("app")
		pg := conn.Config.PostgresEnvConfig
		Expect(pg.GetDatabase()).To(Equal("app_production"))
		Expect(pg.Superuser).To(BeNil())
	})

	It("gives the migration job superuser credentials", func() {
		conn := addonConnTo("migrate")
		pg := conn.Config.PostgresEnvConfig
		Expect(pg.GetDatabase()).To(Equal("app_production"))
		Expect(pg.GetSuperuser()).To(BeTrue())
	})

	It("templates the same DATABASE_URL mapping for both", func() {
		for _, target := range []string{"app", "migrate"} {
			conn := addonConnTo(target)
			Expect(conn.Mappings).To(HaveLen(1))
			m := conn.Mappings[0]
			Expect(m.Target.GetName()).To(Equal("DATABASE_URL"))
			Expect(m.Value.GetTemplate()).To(ContainSubstring("{{ username }}:{{ password }}"))
			Expect(*m.Value.Values).To(HaveKey("password"))
		}
	})
})
