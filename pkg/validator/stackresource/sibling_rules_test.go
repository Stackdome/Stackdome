package stackresource

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

var _ = Describe("sibling rules", func() {
	It("duplicate resource name", func() {
		r := validImageResource() // name "web"
		siblings := []*models.StackResource{{Name: "web"}}
		errs := validateSiblingRules(r, siblings)
		Expect(codes(errs)).To(HaveKey(errors.VErrResourceNameDuplicate))
	})

	It("unknown dependency", func() {
		r := validImageResource()
		r.DependsOn = models.Dependencies{"redis"}
		errs := validateSiblingRules(r, nil)
		Expect(codes(errs)).To(HaveKey(errors.VErrDependencyUnknown))
	})

	It("dependency cycle", func() {
		r := validImageResource()
		r.DependsOn = models.Dependencies{"api"}
		siblings := []*models.StackResource{
			{Name: "api", DependsOn: models.Dependencies{"db"}},
			{Name: "db", DependsOn: models.Dependencies{"web"}}, // db -> web -> api -> db
		}
		errs := validateSiblingRules(r, siblings)
		Expect(codes(errs)).To(HaveKey(errors.VErrDependencyCycle))
	})

	It("self dependency is not also reported as a cycle", func() {
		r := validImageResource()
		r.DependsOn = models.Dependencies{r.Name}
		errs := validateSiblingRules(r, nil)
		Expect(codes(errs)).NotTo(HaveKey(errors.VErrDependencyCycle))
	})

	It("duplicate subdomain prefix across stack", func() {
		r := validImageResource()
		r.Ports[0].SubdomainPrefix = "app"
		siblings := []*models.StackResource{{
			Name:  "other",
			Ports: models.Ports{{Name: "http", Number: 80, Protocol: "http", ExposedToPublic: true, SubdomainPrefix: "app"}},
		}}
		errs := validateSiblingRules(r, siblings)
		Expect(codes(errs)).To(HaveKey(errors.VErrSubdomainDuplicate))
	})

	It("clean graph passes", func() {
		r := validImageResource()
		r.DependsOn = models.Dependencies{"db"}
		siblings := []*models.StackResource{{Name: "db"}}
		errs := validateSiblingRules(r, siblings)
		Expect(errs).To(HaveLen(0))
	})
})
