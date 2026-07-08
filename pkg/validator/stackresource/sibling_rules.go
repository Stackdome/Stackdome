package stackresource

import (
	"fmt"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

// validateSiblingRules checks constraints spanning the whole stack: name
// uniqueness, depends_on target existence and acyclicity, and public
// subdomain uniqueness. Siblings must not include the resource itself.
func validateSiblingRules(resource *models.StackResource, siblings []*models.StackResource) []errors.FieldError {
	var errs []errors.FieldError

	siblingNames := make(map[string]*models.StackResource, len(siblings))
	for _, s := range siblings {
		siblingNames[s.Name] = s
	}

	if _, exists := siblingNames[resource.Name]; exists {
		errs = append(errs, fieldErr("name", errors.VErrResourceNameDuplicate,
			"a resource named '%s' already exists in this stack", resource.Name))
	}

	for i, dep := range resource.DependsOn {
		if dep == resource.Name {
			continue // reported by input rules
		}
		if _, ok := siblingNames[dep]; !ok {
			errs = append(errs, fieldErr(fmt.Sprintf("depends_on[%d]", i), errors.VErrDependencyUnknown,
				"resource '%s' does not exist in stack", dep))
		}
	}

	if hasDependencyCycle(resource, siblings) {
		errs = append(errs, fieldErr("depends_on", errors.VErrDependencyCycle,
			"depends_on introduces a dependency cycle"))
	}

	errs = append(errs, validateSubdomainUniqueness(resource, siblings)...)
	return errs
}

// hasDependencyCycle runs DFS over the stack's dependency graph including the
// candidate resource. A cycle means the resources gate on each other forever
// in the cluster.
func hasDependencyCycle(resource *models.StackResource, siblings []*models.StackResource) bool {
	graph := make(map[string][]string, len(siblings)+1)
	for _, s := range siblings {
		graph[s.Name] = withoutSelfEdge(s.Name, s.DependsOn)
	}
	graph[resource.Name] = withoutSelfEdge(resource.Name, resource.DependsOn)

	const (
		inStack = 1
		done    = 2
	)
	state := map[string]int{}
	var visit func(name string) bool
	visit = func(name string) bool {
		switch state[name] {
		case inStack:
			return true
		case done:
			return false
		}
		state[name] = inStack
		for _, dep := range graph[name] {
			if _, ok := graph[dep]; !ok {
				continue // unknown dep reported separately
			}
			if visit(dep) {
				return true
			}
		}
		state[name] = done
		return false
	}
	for name := range graph {
		if visit(name) {
			return true
		}
	}
	return false
}

// withoutSelfEdge drops a self-reference from deps. Self-dependencies are
// already reported by input rules as VErrDependencySelf and must not also
// surface here as a cycle.
func withoutSelfEdge(name string, deps []string) []string {
	out := make([]string, 0, len(deps))
	for _, d := range deps {
		if d != name {
			out = append(out, d)
		}
	}
	return out
}

func validateSubdomainUniqueness(resource *models.StackResource, siblings []*models.StackResource) []errors.FieldError {
	var errs []errors.FieldError
	taken := map[string]string{} // prefix -> owning resource
	for _, s := range siblings {
		for _, p := range s.Ports {
			if p.ExposedToPublic && p.SubdomainPrefix != "" {
				taken[p.SubdomainPrefix] = s.Name
			}
		}
	}
	seenLocal := map[string]bool{}
	for i, p := range resource.Ports {
		if !p.ExposedToPublic || p.SubdomainPrefix == "" {
			continue
		}
		f := fmt.Sprintf("ports[%d].subdomain_prefix", i)
		if owner, ok := taken[p.SubdomainPrefix]; ok {
			errs = append(errs, fieldErr(f, errors.VErrSubdomainDuplicate,
				"subdomain prefix '%s' is already used by resource '%s'", p.SubdomainPrefix, owner))
		}
		if seenLocal[p.SubdomainPrefix] {
			errs = append(errs, fieldErr(f, errors.VErrSubdomainDuplicate,
				"subdomain prefix '%s' is used twice on this resource", p.SubdomainPrefix))
		}
		seenLocal[p.SubdomainPrefix] = true
	}
	return errs
}
