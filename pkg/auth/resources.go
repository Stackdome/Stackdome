package auth

import (
	"slices"
	"strings"
)

const (
	ResourceStacks          = "stacks"
	ResourceSecrets         = "secrets"
	ResourceVolumes         = "volumes"
	ResourceAddonsPostgres  = "addons/postgres"
	ResourceClusters        = "clusters"
	ResourceImageRegistries = "image-registries"
	ResourceOrgs            = "orgs"
	ResourceTeams           = "teams"
	ResourceObjectStores    = "object-stores"
	ResourceUsers           = "users"
	ResourceWorkspaceUsers  = "workspace-users"
	ResourceImageBuilds     = "image-builds"
	ResourceAddons          = "addons"
	ResourceDomains         = "domains"
	ResourceInvites         = "invites"
	ResourcePreviewConfigs  = "preview-configs"
	ResourcePreviewStacks   = "preview-stacks"
)

const (
	ActionRead   = "read"
	ActionWrite  = "write"
	ActionDelete = "delete"
	ActionCreate = "create"
	ActionList   = "list"
	ActionLogs   = "logs"
	ActionExec   = "exec"

	ScopeFullAccess = "*:*"
)

type ResourceType struct {
	Name    string
	Parent  string // Optional parent resource for hierarchical permissions (e.g. "addons" is parent of "addons/postgres")
	Actions []string
}

var ResourceTypes = []ResourceType{
	{Name: ResourceStacks, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate, ActionLogs, ActionExec}},
	{Name: ResourceSecrets, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate}},
	{Name: ResourceVolumes, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate}},
	{Name: ResourceClusters, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate}},
	{Name: ResourceImageRegistries, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate}},
	{Name: ResourceOrgs, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate}},
	{Name: ResourceAddonsPostgres, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate, ActionLogs, ActionExec}, Parent: ResourceAddons},
	{Name: ResourceObjectStores, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate}},
	{Name: ResourceUsers, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate}},
	{Name: ResourceImageBuilds, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate}},
	{Name: ResourceAddons, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate, ActionLogs, ActionExec}},
	{Name: ResourceDomains, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate}},
	{Name: ResourceTeams, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate}},
	{Name: ResourceInvites, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate}},
	{Name: ResourcePreviewConfigs, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate}},
	{Name: ResourcePreviewStacks, Actions: []string{ActionRead, ActionWrite, ActionDelete, ActionList, ActionCreate}},
}

// ValidateScope checks whether a scope string is valid against the resource type registry.
// Scopes use "resource:action" format. Wildcards are supported: "*:*" grants everything,
// "stacks:*" grants all actions on stacks, and "addons:*" covers child resources like "addons/postgres".
func ValidateScope(scope string) bool {
	if scope == ScopeFullAccess {
		return true
	}
	parts := splitScope(scope)
	if parts == nil {
		return false
	}
	resource, action := parts[0], parts[1]
	for _, rt := range ResourceTypes {
		if rt.Name == resource || (rt.Parent != "" && rt.Parent == resource) {
			if action == "*" {
				return true
			}
			return slices.Contains(rt.Actions, action)
		}
	}
	return false
}

// ScopeCovers returns true if grantedScope permits the requested resource and action.
// Supports wildcard matching: "stacks:*" covers "stacks:read", and "addons:*" covers "addons/postgres:write".
func ScopeCovers(grantedScope, requestedResource, requestedAction string) bool {
	if grantedScope == ScopeFullAccess {
		return true
	}
	parts := splitScope(grantedScope)
	if parts == nil {
		return false
	}
	scopeResource, scopeAction := parts[0], parts[1]

	resourceMatch := scopeResource == requestedResource
	// If resource doesn't match directly, check if it's a parent resource that covers the requested resource
	// Ex: granted scope can be "addons:*" which should cover requested resource "addons/postgres"
	if !resourceMatch {
		// Go through resource types to find if any has scopeResource as parent of requestedResource.
		for _, rt := range ResourceTypes {
			if rt.Name == requestedResource && rt.Parent == scopeResource {
				resourceMatch = true
				break
			}
		}
	}
	if !resourceMatch {
		return false
	}
	return scopeAction == "*" || scopeAction == requestedAction
}

// splitScope splits a scope string on the last colon, returning [resource, action].
// Examples:
//
//	"stacks:read"          → ["stacks", "read"]
//	"addons/postgres:*"    → ["addons/postgres", "*"]
//	"*:*"                  → ["*", "*"]
//	"stacks"               → nil (no colon, invalid)
func splitScope(scope string) []string {
	parts := strings.Split(scope, ":")
	if len(parts) != 2 {
		return nil
	}
	return parts
}
